package web

import (
	"fmt"
	"net/http"
	"strings"

	"brooce/heartbeat"
	"brooce/listing"
	"brooce/task"

	redis "gopkg.in/redis.v3"
)

type mainpageOutputType struct {
	Queues         map[string]*listQueueType
	RunningJobs    []*task.Task
	RunningWorkers []*heartbeat.HeartbeatTemplateType
	TotalThreads   int
	AliveWorkers   int
}

func mainpageHandler(req *http.Request, rep *httpReply) (err error) {
	output := &mainpageOutputType{}

	output.RunningJobs, err = listing.RunningJobs()
	if err != nil {
		return
	}
	output.RunningWorkers, output.AliveWorkers, err = listing.RunningWorkers()
	if err != nil {
		return
	}
	output.Queues, err = listQueues(output.RunningWorkers)
	if err != nil {
		return
	}

	for _, worker := range output.RunningWorkers {
		output.TotalThreads += worker.TotalThreads()
	}

	err = templates.ExecuteTemplate(rep, "mainpage", output)
	return
}

type listQueueType struct {
	QueueName     string
	Pending       int64
	Running       int
	Done          int64
	Failed        int64
	Delayed       int64
	pendingResult *redis.IntCmd
	runningResult *redis.StringSliceCmd
	doneResult    *redis.IntCmd
	failedResult  *redis.IntCmd
	delayedResult *redis.IntCmd
}

func listQueues(runningWorkers []*heartbeat.HeartbeatTemplateType) (list map[string]*listQueueType, err error) {
	list = map[string]*listQueueType{}
	var results []string
	results, err = redisClient.Keys(redisHeader + ":queue:*").Result()
	if err != nil {
		return
	}

	for _, result := range results {
		parts := strings.Split(result, ":")
		if len(parts) < 3 {
			continue
		}

		list[parts[2]] = &listQueueType{QueueName: parts[2]}
	}

	for _, worker := range runningWorkers {
		for queueName, _ := range worker.Queues {
			list[queueName] = &listQueueType{QueueName: queueName}
		}
	}

	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for _, queue := range list {
			queue.pendingResult = pipe.LLen(fmt.Sprintf("%s:queue:%s:pending", redisHeader, queue.QueueName))
			queue.runningResult = pipe.Keys(fmt.Sprintf("%s:queue:%s:working:*", redisHeader, queue.QueueName))
			queue.doneResult = pipe.LLen(fmt.Sprintf("%s:queue:%s:done", redisHeader, queue.QueueName))
			queue.failedResult = pipe.LLen(fmt.Sprintf("%s:queue:%s:failed", redisHeader, queue.QueueName))
			queue.delayedResult = pipe.LLen(fmt.Sprintf("%s:queue:%s:delayed", redisHeader, queue.QueueName))
		}
		return nil
	})
	if err != nil {
		return
	}

	for _, queue := range list {
		queue.Pending = queue.pendingResult.Val()
		queue.Running = len(queue.runningResult.Val())
		queue.Done = queue.doneResult.Val()
		queue.Failed = queue.failedResult.Val()
		queue.Delayed = queue.delayedResult.Val()
	}

	return
}
