package listing

import (
	"encoding/json"

	"brooce/heartbeat"
)

func RunningWorkers() (workers []*heartbeat.HeartbeatType, err error) {
	redisKey := redisHeader + ":workerprocs"
	var results map[string]string
	results, err = redisClient.HGetAll(redisKey).Result()
	if err != nil || len(results) == 0 {
		return
	}

	for hKey, str := range results {
		worker := &heartbeat.HeartbeatType{}
		err = json.Unmarshal([]byte(str), worker)

		if err != nil || worker.HeartbeatTooOld() || worker.IsLocalZombie() {
			err = redisClient.HDel(redisKey, hKey).Err()
			if err != nil {
				return
			}

			continue
		}

		workers = append(workers, worker)
	}

	return
}
