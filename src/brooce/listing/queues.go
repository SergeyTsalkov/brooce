package listing

import (
	"fmt"

	"brooce/heartbeat"

	redis "gopkg.in/redis.v6"
)

type QueueInfoType struct {
	Name          string
	Workers       int64
	Pending       int64
	Running       int64
	Done          int64
	Failed        int64
	Delayed       int64
	pendingResult *redis.IntCmd
	//runningResult *redis.StringSliceCmd
	doneResult    *redis.IntCmd
	failedResult  *redis.IntCmd
	delayedResult *redis.IntCmd
}

func Queues(short bool) (queueHash map[string]*QueueInfoType, err error) {
	var workers []*heartbeat.HeartbeatType
	workers, err = RunningWorkers()
	if err != nil || len(workers) == 0 {
		return
	}

	queueHash = map[string]*QueueInfoType{}
	for _, worker := range workers {
		for _, queue := range worker.Queues {

			if queueHash[queue.Name] == nil {
				queueHash[queue.Name] = &QueueInfoType{}
			}

			queueInfo := queueHash[queue.Name]
			queueInfo.Name = queue.Name
			queueInfo.Workers += int64(queue.Workers)
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

	for _, queue := range queueHash {
		queue.Pending = queue.pendingResult.Val()
		queue.Done = queue.doneResult.Val()
		queue.Failed = queue.failedResult.Val()
		queue.Delayed = queue.delayedResult.Val()
	}

	return
}

/*
func QueueList(short bool) (queueList []*QueueInfoType, err error) {
	var queueHash map[string]*QueueInfoType
	queueHash, err = QueueHash(short)
	if err != nil {
		return
	}

	for _, queue := range queueHash {
		queueList = append(queueList, queue)
	}

	sort.Slice(queueList, func(i, j int) bool {
		return queueList[i].Name < queueList[j].Name
	})

	return
}
*/
