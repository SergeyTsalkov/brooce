package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"brooce/config"
	"brooce/cron/cronsched"
	"brooce/heartbeat"
	"brooce/lock"
	"brooce/prune"
	myredis "brooce/redis"
	"brooce/requeue"
	"brooce/runnabletask"
	"brooce/suicide"
	tasklib "brooce/task"
	"brooce/web"

	daemon "github.com/sevlyar/go-daemon"
	redis "gopkg.in/redis.v5"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

var daemonizeOpt = flag.Bool("daemonize", false, "Detach and run in the background!")
var helpOpt = flag.Bool("help", false, "Show these options!")

func setup() {
	heartbeat.Start()
	web.Start()
	cronsched.Start()
	prune.Start()
	requeue.Start()
	suicide.Start()
	lock.Start()
}

func main() {
	flag.Parse()
	if *helpOpt {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *daemonizeOpt {
		context := &daemon.Context{
			LogFileName: filepath.Join(config.BrooceDir, "brooce.log"),
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

	setup()

	strQueueList := []string{}

	for _, q := range config.Config.Queues {
		for i := 0; i < q.Workers; i++ {
			go runner(q.Name, i)
		}

		strQueueList = append(strQueueList, fmt.Sprintf("%v (x%v)", q.Name, q.Workers))
	}

	if len(config.Config.Queues) > 0 {
		log.Println("Started with queues:", strings.Join(strQueueList, ", "))
	} else {
		log.Println("Started with NO queues! We won't be doing any jobs!")
	}

	select {} //sleep forever!
}

func runner(queue string, ct int) {
	threadName := fmt.Sprintf("%v-%v-%v", config.Config.ProcName, queue, ct)

	pendingList := fmt.Sprintf("%s:queue:%s:pending", redisHeader, queue)
	workingList := fmt.Sprintf("%s:queue:%s:working:%s", redisHeader, queue, threadName)
	doneList := fmt.Sprintf("%s:queue:%s:done", redisHeader, queue)
	failedList := fmt.Sprintf("%s:queue:%s:failed", redisHeader, queue)
	delayedList := fmt.Sprintf("%s:queue:%s:delayed", redisHeader, queue)

	var threadOutputLog *os.File
	if config.Config.FileOutputLog.Enable {
		var err error
		filename := filepath.Join(config.BrooceDir, fmt.Sprintf("%s-%s-%d.log", config.Config.ClusterName, queue, ct))
		threadOutputLog, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalln("Unable to open logfile", filename, "for writing! Error was", err)
		}
		defer threadOutputLog.Close()
	}

	for {
		taskStr, err := redisClient.BRPopLPush(pendingList, workingList, 15*time.Second).Result()
		if err != nil {
			if err != redis.Nil {
				log.Println("redis error while running BRPOPLPUSH:", err)
			}
			continue
		}

		var exitCode int
		task, err := tasklib.NewFromJson(taskStr, config.Config.LocalOptionsForQueue(queue))
		if err != nil {
			log.Println("Failed to decode task:", err)
		} else {
			task.RedisKey = workingList
			rTask := &runnabletask.RunnableTask{
				Task:       task,
				FileWriter: threadOutputLog,
			}
			suicide.ThreadIsWorking(threadName)
			exitCode, err = rTask.Run()
			suicide.ThreadIsWaiting(threadName)

			if err != nil {
				log.Printf("Error in task %v: %v", rTask.Id, err)
			}
		}

		redisClient.Pipelined(func(pipe *redis.Pipeline) error {
			result := "failed"

			if err != nil {
				result = "failed"
			} else if exitCode == 0 {
				result = "done"
			} else if exitCode == 75 {
				// Unix standard "temp fail" code
				result = "delayed"
			}

			switch result {
			case "done":
				if !config.Config.JobResults.DropDone {
					pipe.LPush(doneList, task.Json())
				}

				if config.Config.RedisOutputLog.DropDone {
					pipe.Del(fmt.Sprintf("%s:jobs:%s:log", redisHeader, task.Id))
				}

			case "failed":
				if !config.Config.JobResults.DropFailed {
					pipe.LPush(failedList, task.Json())
				}

				if config.Config.RedisOutputLog.DropFailed {
					pipe.Del(fmt.Sprintf("%s:jobs:%s:log", redisHeader, task.Id))
				}
			case "delayed":
				pipe.LPush(delayedList, task.Json())
			}

			pipe.RPop(workingList)
			return nil
		})

		// workingList should be empty by this point
		// if it's not, something went wrong earlier
		err = myredis.FlushList(workingList, failedList)
		if err != nil {
			log.Println("Error while flushing", workingList, "to", failedList, ":", err)
		}

	}
}
