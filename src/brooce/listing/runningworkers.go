package listing

import (
	"encoding/json"

	"brooce/heartbeat"

	redis "gopkg.in/redis.v3"
	"time"
)

func RunningWorkers() (workers []*heartbeat.HeartbeatTemplateType, aliveWorkers int, err error) {
	var keys []string
	keys, err = redisClient.Keys(redisHeader + ":workerprocs:*").Result()
	if err != nil {
		return
	}

	var heartbeatStrs []*redis.StringCmd
	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for _, key := range keys {
			result := pipe.Get(key)
			heartbeatStrs = append(heartbeatStrs, result)
		}
		return nil
	})
	if err != nil {
		return
	}

	for _, str := range heartbeatStrs {
		worker := &heartbeat.HeartbeatTemplateType{}
		err = json.Unmarshal([]byte(str.Val()), worker)
		if err != nil {
			return
		}

		workerTS := time.Unix(int64(worker.TS), 0)
		worker.PrettyTS = workerTS.Format(time.RFC3339)

		switch IsAlive(workerTS) {
		case -1:
			worker.StatusColor = "red"
		case 0:
			worker.StatusColor = "yellow"
		case 1:
			worker.StatusColor = "green"
			aliveWorkers = aliveWorkers + 1
		default:
			worker.StatusColor = "grey"
		}

		workers = append(workers, worker)
	}

	return
}

func IsAlive(workerTS time.Time) int {
	currentTS := time.Now().Unix()

	if currentTS > workerTS.Add(heartbeat.AssumeDeadAfter).Unix() {
		return -1
	} else if currentTS < workerTS.Add(heartbeat.AssumeDeadAfter).Unix() && currentTS > workerTS.Add(heartbeat.HeartbeatEvery).Unix() {
		return 0
	} else if currentTS <= workerTS.Add(heartbeat.HeartbeatEvery).Unix() {
		return 1
	} else {
		return -11
	}
}