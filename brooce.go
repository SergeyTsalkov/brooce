package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"brooce/config"
	"brooce/cron/cronsched"
	"brooce/heartbeat"
	"brooce/lock"
	"brooce/prune"
	myredis "brooce/redis"
	"brooce/requeue"
	"brooce/runnabletask"
	mysignals "brooce/signals"
	"brooce/suicide"
	tasklib "brooce/task"
	"brooce/web"

	"github.com/go-redis/redis"
	daemon "github.com/sevlyar/go-daemon"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName
var threadWg sync.WaitGroup

var daemonizeOpt = flag.Bool("daemonize", false, "Detach and run in the background!")
var helpOpt = flag.Bool("help", false, "Show these options!")

func main() {
	flag.Parse()
	if *helpOpt {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *daemonizeOpt {
		context := &daemon.Context{
			LogFileName: filepath.Join(config.BrooceLogDir, "brooce.log"),
			LogFilePerm: 0644,
		}
		child, err := context.Reborn()
		if err != nil {
			log.Fatalln("Daemonize error:", err)
		}

		if child != nil {
			log.Println("Starting brooce in the background..")
			os.Exit(0)
		} else {
			defer context.Release()
		}
	}

	heartbeat.Start()
	web.Start()
	cronsched.Start()
	prune.Start()
	requeue.Start()
	suicide.Start()
	lock.Start()
	mysignals.Start()

	for _, thread := range config.Threads {
		threadWg.Add(1)
		go runner(thread)
	}

	if len(config.Threads) > 0 {
		log.Println("Started with queues:", config.ThreadString)
	} else {
		log.Println("Started with NO queues! We won't be doing any jobs!")
	}

	mysignals.WaitForShutdownRequest()
	log.Println("Shutdown requested! Waiting for all threads to finish (repeat signal to skip)..")
	threadWg.Wait()
	log.Println("Exiting..")
}

func runner(thread config.ThreadType) {
	var threadOutputLog *os.File
	if config.Config.FileOutputLog.Enable {
		var err error
		filename := filepath.Join(config.BrooceLogDir, fmt.Sprintf("%s-%s-%d.log", config.Config.ClusterName, thread.Queue, thread.Id))
		threadOutputLog, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalln("Unable to open logfile", filename, "for writing! Error was", err)
		}
		defer threadOutputLog.Close()
	}

	for {
		if mysignals.WasShutdownRequested() {
			threadWg.Done()
			return
		}

		taskStr, err := redisClient.BRPopLPush(thread.PendingList(), thread.WorkingList(), 15*time.Second).Result()
		if err != nil {
			if err != redis.Nil {
				log.Println("redis error while running BRPOPLPUSH:", err)
			}
			continue
		}

		// thread.WorkingList() should have 1 item now
		// if it has more, something went wrong!
		length := redisClient.LLen(thread.WorkingList())
		if length.Err() != nil {
			log.Println("Error while checking length of", thread.WorkingList(), ":", err)
		}
		if length.Val() != 1 {
			log.Println(thread.WorkingList(), "should have length 1 but has length", length.Val(), "! It'll be flushed to", thread.PendingList())

			err = myredis.FlushList(thread.WorkingList(), thread.PendingList())
			if err != nil {
				log.Println("Error while flushing", thread.WorkingList(), "to", thread.PendingList(), ":", err)
			}
			continue
		}

		var exitCode int
		task, err := tasklib.NewFromJson(taskStr, thread.Queue)
		if err != nil {
			log.Println("Failed to decode task:", err)
		} else {
			task.RedisKey = thread.WorkingList()
			rTask := &runnabletask.RunnableTask{
				Task:       task,
				FileWriter: threadOutputLog,
			}
			suicide.ThreadIsWorking(thread.Name)
			exitCode, err = rTask.Run()
			suicide.ThreadIsWaiting(thread.Name)

			if err != nil && !strings.HasPrefix(err.Error(), "timeout after") && !strings.HasPrefix(err.Error(), "exit status") {
				log.Printf("Error in task %v: %v", rTask.Id, err)
			}
		}

		_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {

			// Unix standard "temp fail" code
			if exitCode == 75 {
				// DELAYED
				if !task.KillOnDelay() {
					pipe.LPush(thread.DelayedList(), task.Json())
				}

			} else if (err != nil || exitCode != 0) && !task.NoFail() {
				// FAILED
				if task.MaxTries() > task.Tried {
					pipe.LPush(thread.DelayedList(), task.Json())
				} else {
					if task.RedisLogFailedExpireAfter() > 0 {
						pipe.Expire(task.LogKey(), time.Duration(task.RedisLogFailedExpireAfter())*time.Second)
					}

					if !task.DropOnFail() {
						pipe.LPush(thread.FailedList(), task.Json())
					}

					if task.NoRedisLogOnFail() && task.LogKey() != "" {
						pipe.Del(task.LogKey())
					}
				}

			} else {
				// DONE

				if !task.DropOnSuccess() {
					pipe.LPush(thread.DoneList(), task.Json())
				}

				if task.NoRedisLogOnSuccess() && task.LogKey() != "" {
					pipe.Del(task.LogKey())
				}
			}

			pipe.LPop(thread.WorkingList())
			return nil
		})

		if err != nil {
			log.Println("Error while pipelining job from", thread.WorkingList(), ":", err)
		}

	}
}
