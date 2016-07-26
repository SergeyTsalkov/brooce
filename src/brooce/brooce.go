package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tasklib "brooce/task"

	redis "gopkg.in/redis.v3"
)

var queueWg = new(sync.WaitGroup)

var redisHeader = "brooce"
var heartbeatKey = redisHeader + ":workerprocs"

func main() {
	init_config()

	// need to send a single heartbeat FOR SURE before we grab a job!
	heartbeat()
	go heartbeater()
	go jobpruner()
	go cronner()
	go suicider()

	if config.Queues == nil {
		logger.Fatalln("The queues hash was not configured in the ~/.brooce config file!")
	}

	threadid := 1

	strQueueList := []string{}

	for queue, ct := range config.Queues {
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
	threadName := fmt.Sprintf("%v-%v", myProcName, threadid)

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
