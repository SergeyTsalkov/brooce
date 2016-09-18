package listing

import (
	"encoding/json"

	"brooce/heartbeat"

	redis "gopkg.in/redis.v3"
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

		if heartbeat.IsAlive(worker) == 1 {
			aliveWorkers = aliveWorkers + 1
		}
		workers = append(workers, worker)
	}

	return
}
