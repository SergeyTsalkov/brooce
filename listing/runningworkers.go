package listing

import (
	"encoding/json"
	"sort"

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

	sort.Slice(workers, func(i, j int) bool {
		if workers[i].Hostname == workers[j].Hostname {
			return workers[i].ProcName < workers[j].ProcName
		} else {
			return workers[i].Hostname < workers[j].Hostname
		}
	})

	return
}
