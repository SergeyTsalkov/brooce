package listing

import (
	"fmt"

	"brooce/config"
	"brooce/heartbeat"
	"brooce/task"

	"github.com/go-redis/redis"
)

type QueueInfoType struct {
	Name          string
	Threads       int64
	Pending       int64
	Running       int64
	Done          int64
	Failed        int64
	Delayed       int64
	pendingResult *redis.IntCmd
	doneResult    *redis.IntCmd
	failedResult  *redis.IntCmd
	delayedResult *redis.IntCmd
}

func (q *QueueInfoType) PendingList() string {
	return fmt.Sprintf("%s:queue:%s:pending", config.Config.ClusterName, q.Name)
}

func (q *QueueInfoType) DoneList() string {
	return fmt.Sprintf("%s:queue:%s:done", config.Config.ClusterName, q.Name)
}

func (q *QueueInfoType) FailedList() string {
	return fmt.Sprintf("%s:queue:%s:failed", config.Config.ClusterName, q.Name)
}

func (q *QueueInfoType) DelayedList() string {
	return fmt.Sprintf("%s:queue:%s:delayed", config.Config.ClusterName, q.Name)
}

// global list of queues, including those on other machines in our cluster
func Queues(short bool) (queueHash map[string]*QueueInfoType, err error) {
	var workers []*heartbeat.HeartbeatType
	workers, err = RunningWorkers()
	if err != nil || len(workers) == 0 {
		return
	}

	queueHash = map[string]*QueueInfoType{}
	for _, worker := range workers {
		for _, thread := range worker.Threads {

			if queueHash[thread.Queue] == nil {
				queueHash[thread.Queue] = &QueueInfoType{}
			}

			queueInfo := queueHash[thread.Queue]
			queueInfo.Name = thread.Queue
			queueInfo.Threads += 1
		}
	}

	if len(queueHash) == 0 || short {
		return
	}

	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for name, queue := range queueHash {
			queue.pendingResult = pipe.LLen(fmt.Sprintf("%s:queue:%s:pending", redisHeader, name))
			queue.doneResult = pipe.LLen(fmt.Sprintf("%s:queue:%s:done", redisHeader, name))
			queue.failedResult = pipe.LLen(fmt.Sprintf("%s:queue:%s:failed", redisHeader, name))
			queue.delayedResult = pipe.LLen(fmt.Sprintf("%s:queue:%s:delayed", redisHeader, name))
		}
		return nil
	})
	if err != nil {
		return
	}

	var jobs []*task.Task
	jobs, err = RunningJobs(true)
	if err != nil {
		return
	}
	for _, job := range jobs {
		queueName := job.QueueName()
		if queueInfo := queueHash[queueName]; queueInfo != nil {
			queueInfo.Running += 1
		}
	}

	for _, queue := range queueHash {
		queue.Pending = queue.pendingResult.Val()
		queue.Done = queue.doneResult.Val()
		queue.Failed = queue.failedResult.Val()
		queue.Delayed = queue.delayedResult.Val()
	}

	return
}
