package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"brooce/config"
	"brooce/cron"
	"brooce/heartbeat"
	"brooce/lock"
	loggerlib "brooce/logger"
	"brooce/prune"
	myredis "brooce/redis"
	"brooce/requeue"
	"brooce/suicide"
	tasklib "brooce/task"
	"brooce/web"

	redis "gopkg.in/redis.v3"
)

var redisClient = myredis.Get()
var logger = loggerlib.Logger
var redisHeader = config.Config.ClusterName

func setup() {
	heartbeat.Start()
	web.Start()
	cron.Start()
	prune.Start()
	requeue.Start()
	suicide.Start()
	lock.Start()
}

func main() {
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

		exitCode := 256
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
			switch exitCode {
			case 0:
				pipe.LPush(doneList, task.Json())
			case 75: // Unix standard "temp fail" code
				pipe.LPush(delayedList, task.Json())
			default:
				pipe.LPush(failedList, task.Json())
			}

			pipe.RPop(workingList)
			return nil
		})

	}
}
