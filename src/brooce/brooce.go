package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"brooce/config"
	"brooce/heartbeat"
	loggerlib "brooce/logger"
	"brooce/myip"
	myredis "brooce/redis"
	tasklib "brooce/task"
	"brooce/web"

	redis "gopkg.in/redis.v3"
)

var redisClient = myredis.Get()
var myProcName string
var logger = loggerlib.Logger
var queueWg = new(sync.WaitGroup)

var redisHeader = config.Config.ClusterName
var heartbeatKey = redisHeader + ":workerprocs"
var publicIP = myip.PublicIPv4()

func setup() {
	web.Start()
	heartbeat.Start()
	go jobpruner()
	go cronner()
	go suicider()
}

func main() {
	setup()

	threadid := 1
	strQueueList := []string{}

	for queue, ct := range config.Config.Queues {
		for i := 0; i < ct; i++ {
			queueWg.Add(1)
			go runner(queue, threadid)
			threadid++
		}

		strQueueList = append(strQueueList, fmt.Sprintf("%v (x%v)", queue, ct))
		go autoRequeueDelayed(queue)
	}

	logger.Println("Started with queues:", strings.Join(strQueueList, ", "))
	queueWg.Wait()
}

func runner(queue string, threadid int) {
	threadName := fmt.Sprintf("%v-%v", config.Config.ProcName, threadid)

	pendingList := strings.Join([]string{redisHeader, "queue", queue, "pending"}, ":")
	workingList := strings.Join([]string{redisHeader, "queue", queue, "working", threadName}, ":")
	doneList := strings.Join([]string{redisHeader, "queue", queue, "done"}, ":")
	failedList := strings.Join([]string{redisHeader, "queue", queue, "failed"}, ":")
	delayedList := strings.Join([]string{redisHeader, "queue", queue, "delayed"}, ":")

	for {
		taskStr, err := redisClient.BRPopLPush(pendingList, workingList, 15*time.Second).Result()
		if err != nil {
			continue
		}

		task, err := tasklib.NewFromJson(taskStr)
		if err != nil {
			fmt.Println("Failed to decode task:", err)
			continue
		}

		announceStatusWorking(threadid)
		exitCode := (&runnableTask{task}).Run()
		announceStatusWaiting(threadid)

		redisClient.Pipelined(func(pipe *redis.Pipeline) error {
			switch exitCode {
			case 0:
				pipe.LPush(doneList, task.Json())
			case 75: // Unix standard "temp fail" code
				pipe.LPush(delayedList, task.Json())
			default:
				pipe.LPush(failedList, task.Json())
			}

			// we're done who cares about this job
			_ = pipe.RPop(workingList)

			return nil
		})

	}

	queueWg.Done()
}

func sleepUntil00() {
	now := time.Now().Unix()
	last_minute := now - now%60
	next_minute := last_minute + 60
	sleep_for := next_minute - now

	time.Sleep(time.Duration(sleep_for) * time.Second)
}
