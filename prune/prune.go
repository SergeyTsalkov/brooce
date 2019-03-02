package prune

import (
	"fmt"
	"log"
	"strings"
	"time"

	"brooce/config"
	"brooce/heartbeat"
	"brooce/listing"
	myredis "brooce/redis"
	"brooce/task"

	"github.com/go-redis/redis"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

func Start() {
	go func() {
		for {
			err := prunejobs()
			if err != nil {
				log.Println("Error while pruning jobs:", err)
			}

			err = prunequeues()
			if err != nil {
				log.Println("Error while pruning queues:", err)
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

func prunequeues() error {
	queueHash, err := listing.Queues(true)
	if err != nil {
		return err
	}

	for name, _ := range queueHash {
		opts := config.Config.DeepJobOptions(name)

		err = expireList(fmt.Sprintf("%s:queue:%s:done",   redisHeader, name), opts.RedisListDoneExpireAfter())
		if err != nil {
			return err
		}

		err = expireList(fmt.Sprintf("%s:queue:%s:failed", redisHeader, name), opts.RedisListFailedExpireAfter())
		if err != nil {
			return err
		}
	}

	return nil
}

func expireList(list string, expire int) error {
	var err error
	var taskStr string

	if expire == 0 {
	    return nil
	}

	queueName := task.QueueNameFromRedisKey(list)
	for lastExpired := true; lastExpired; {
		taskStr, err = redisClient.LIndex(list, -1).Result()
		// empty list
		if err == redis.Nil {
			break
		}
		if err != nil {
			return err
		}

		job, err := task.NewFromJson(taskStr, queueName)
		if err != nil {
			return err
		}

		// grab the job...
		if jobHasExpired(job, expire) {
			taskStr, err = redisClient.RPop(list).Result()
			// it's possible for an item to vanish between the LINDEX and RPOP steps -- this is not fatal!
			if err == redis.Nil {
				break
			}
			if err != nil {
				return err
			}

			job, err = task.NewFromJson(taskStr, queueName)
			if err != nil {
				return err
			}

			// ...and recheck for sure, this could be an another job too
			if !jobHasExpired(job, expire) {
				err = redisClient.RPush(list, taskStr).Err()
				if err != nil {
					return err
				}
				lastExpired = false
			}

		} else {
			lastExpired = false
		}
	}

	return nil
}

func jobHasExpired(job *task.Task, expire int) bool {
	if job.EndTime > 0 && job.EndTime < time.Now().Unix() - int64(expire) {
		return true
	}

	return false
}
