package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"brooce/config"
	"brooce/cron/cronsched"
	"brooce/heartbeat"
	"brooce/lock"
	loggerlib "brooce/logger"
	"brooce/prune"
	myredis "brooce/redis"
	"brooce/requeue"
	"brooce/suicide"
	tasklib "brooce/task"
	"brooce/web"

	daemon "github.com/sevlyar/go-daemon"
	redis "gopkg.in/redis.v3"
)

var redisClient = myredis.Get()
var logger = loggerlib.Logger
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
		context := &daemon.Context{}
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

	threadid := 1
	strQueueList := []string{}

	for queue, ct := range config.Config.Queues {
		for i := 0; i < ct; i++ {
			go runner(queue, threadid)
			threadid++
		}

		strQueueList = append(strQueueList, fmt.Sprintf("%v (x%v)", queue, ct))
	}

	if len(config.Config.Queues) > 0 {
		logger.Println("Started with queues:", strings.Join(strQueueList, ", "))
	} else {
		logger.Println("Started with NO queues! We won't be doing any jobs!")
	}

	select {} //sleep forever!
}

func runner(queue string, threadid int) {
	threadName := fmt.Sprintf("%v-%v", config.Config.ProcName, threadid)

	pendingList := fmt.Sprintf("%s:queue:%s:pending", redisHeader, queue)
	workingList := fmt.Sprintf("%s:queue:%s:working:%s", redisHeader, queue, threadName)
	doneList := fmt.Sprintf("%s:queue:%s:done", redisHeader, queue)
	failedList := fmt.Sprintf("%s:queue:%s:failed", redisHeader, queue)
	delayedList := fmt.Sprintf("%s:queue:%s:delayed", redisHeader, queue)

	for {
		taskStr, err := redisClient.BRPopLPush(pendingList, workingList, 15*time.Second).Result()
		if err != nil {
			if err != redis.Nil {
				log.Println("redis error while running BRPOPLPUSH:", err)
			}
			continue
		}

		var exitCode int
		task, err := tasklib.NewFromJson(taskStr)
		if err != nil {
			log.Println("Failed to decode task:", err)
		} else {
			rTask := &runnableTask{
				Task:        task,
				workingList: workingList,
				threadName:  threadName,
				queueName:   queue,
			}
			suicide.ThreadIsWorking(threadid)
			exitCode, err = rTask.Run()
			suicide.ThreadIsWaiting(threadid)

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

	}
}
