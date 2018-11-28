package prune

import (
	"fmt"
	"log"
	"strings"
	"time"

	"brooce/heartbeat"
	"brooce/listing"
	myredis "brooce/redis"
	"brooce/task"
)

func Start() {
	go func() {
		for {
			err := prunejobs()
			if err != nil {
				log.Println("Error while pruning jobs:", err)
			}

			time.Sleep(time.Minute)
		}
	}()
}

func prunejobs() error {
	workers, err := listing.RunningWorkers()
	if err != nil {
		return err
	}

	jobs, err := listing.RunningJobs(false)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if !jobHasWorker(job, workers) {
			deadList := job.RedisKey

			parts := strings.SplitN(deadList, ":", 5)
			if len(parts) < 5 {
				log.Println("Weird working queue found:", deadList)
				continue
			}

			failedList := fmt.Sprintf("%s:queue:%s:failed", parts[0], parts[2])
			log.Println("Pruning dead working queue", deadList, "to", failedList)

			err = myredis.FlushList(deadList, failedList)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func jobHasWorker(job *task.Task, workers []*heartbeat.HeartbeatType) bool {
	for _, worker := range workers {
		if strings.HasPrefix(job.WorkerThreadName(), worker.ProcName+"-") {
			return true
		}
	}

	return false
}
