package prune

import (
	"fmt"
	"log"
	"strings"
	"time"

	"brooce/config"
	myredis "brooce/redis"
	"brooce/heartbeat"
	"encoding/json"
	tasklib "brooce/task"

	"brooce/lock"
)

var redisHeader = config.Config.ClusterName
var heartbeatKey = redisHeader + ":workerprocs"
var redisClient = myredis.Get()

func Start() {
	go func() {
		for {
			prunejobs()
			time.Sleep(time.Minute)
		}
	}()
}

func prunejobs() {
	for _, proc := range deadProcs() {
		pruneProc(proc)
	}
}

func pruneProc(procName string) {
	results, err := redisClient.Keys(redisHeader + ":queue:*:working:" + procName).Result()
	if err != nil {
		return
	}

	for _, result := range results {
		parts := strings.Split(result, ":")
		if len(parts) < 5 {
			continue
		}

		queue := parts[2]
		failedList := strings.Join([]string{redisHeader, "queue", queue, "failed"}, ":")
		log.Println("Pruning dead working queue", result, "to", failedList)

		taskStr, err := redisClient.RPopLPush(result, failedList).Result()
		task, err := tasklib.NewFromJson(taskStr)
		if err != nil {
			log.Println("Failed to decode task:", err)
		} else {
			lock.ReleaseLocks(task.Locks)
		}
	}
}

func beatingProcs() ([]string, error) {
	livingProcs := []string{}

	results, err := redisClient.Keys(heartbeatKey + ":*").Result()
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		var livingProc string
		fmt.Sscanf(result, heartbeatKey+":%s", &livingProc)

		str := myredis.Get().Get(result)

		worker := &heartbeat.HeartbeatTemplateType{}
		err = json.Unmarshal([]byte(str.Val()), worker)
		if err != nil {
			return nil, err
		}

		if heartbeat.IsAlive(worker) != -1 {
			livingProcs = append(livingProcs, livingProc)
		}
	}

	if len(livingProcs) == 0 {
		return nil, fmt.Errorf("Couldn't find any living processes!")
	}
	return livingProcs, nil
}

func workingProcs() []string {
	workingProcs := []string{}

	results, err := redisClient.Keys(redisHeader + ":queue:*:working:*").Result()
	if err != nil {
		return workingProcs
	}

outer:
	for _, result := range results {
		parts := strings.Split(result, ":")
		if len(parts) < 5 {
			continue
		}

		workingProc := parts[4]

		if workingProcParts := strings.Split(workingProc, "-"); len(workingProcParts) == 3 {
			workingProc = strings.Join(workingProcParts[0:2], "-")
		}

		workingProcTaskList, err := myredis.Get().LRange(result, 0, -1).Result()
		if err != nil {
			log.Println(err)
			return workingProcs
		}

		for _, str := range workingProcTaskList {
			worker := &heartbeat.HeartbeatTemplateType{}
			err = json.Unmarshal([]byte(str), worker)
			if err != nil {
				log.Println(err)
				return workingProcs
			}

			//dedup
			for _, alreadyListed := range workingProcs {
				if alreadyListed == workingProc {
					continue outer
				}
			}

			workingProcs = append(workingProcs, workingProc)
		}
	}

	return workingProcs
}

func deadProcs() []string {
	deadProcs := []string{}

	beatingProcs, err := beatingProcs()
	if err != nil {
		return deadProcs
	}

outer:
	for _, working := range workingProcs() {
		for _, beating := range beatingProcs {
			if working == beating {
				continue outer
			}
		}

		deadProcs = append(deadProcs, working)
	}

	return deadProcs
}
