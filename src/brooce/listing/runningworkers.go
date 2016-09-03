package listing

import (
	"encoding/json"

	"brooce/heartbeat"

	redis "gopkg.in/redis.v3"
	"time"
)

func RunningWorkers() (workers []*heartbeat.HeartbeatTemplateType, err error) {
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

		currentTS := time.Now().Unix()

		if currentTS > workerTS.Add(heartbeat.AssumeDeadAfter).Unix() {
			worker.StatusColor = "red"
		} else if currentTS < workerTS.Add(heartbeat.AssumeDeadAfter).Unix() && currentTS > workerTS.Add(heartbeat.HeartbeatEvery).Unix() {
			worker.StatusColor = "yellow"
		} else if currentTS <= workerTS.Add(heartbeat.HeartbeatEvery).Unix() {
			worker.StatusColor = "green"
		} else {
			worker.StatusColor = "grey"
		}

		workers = append(workers, worker)
	}

	return
}
