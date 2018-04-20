package listing

import (
	"fmt"

	"brooce/config"
	"brooce/heartbeat"
	myredis "brooce/redis"
	"brooce/task"

	redis "gopkg.in/redis.v6"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

// SCAN takes about 0.5s per million total items in redis
// we skip it by guessing the possible working list names
// from worker heartbeat data
// this is much faster, but the prune functions still need
// the true list to find any zombie working lists
func RunningJobs(fast bool) (jobs []*task.Task, err error) {
	jobs = []*task.Task{}

	var keys []string

	if fast {
		var workers []*heartbeat.HeartbeatType

		workers, err = RunningWorkers()
		if err != nil {
			return
		}

		for _, worker := range workers {
			for _, thread := range worker.Threads {
				keys = append(keys, thread.WorkingList())
			}
		}

	} else {
		keys, err = myredis.ScanKeys(redisHeader + ":queue:*:working:*")
		if err != nil {
			return
		}
	}

	if len(keys) == 0 {
		return
	}

	values := make([]*redis.StringCmd, len(keys))
	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for i, key := range keys {
			values[i] = pipe.LIndex(key, 0)
		}
		return nil
	})
	// it's possible for an item to vanish between the KEYS and LINDEX steps -- this is not fatal!
	if err == redis.Nil {
		err = nil
	}
	if err != nil {
		return
	}

	for i, value := range values {
		if value.Err() != nil {
			// possible to get a redis.Nil error here if a job vanished between the KEYS and LINDEX steps
			continue
		}
		job, err := task.NewFromJson(value.Val(), config.JobOptions{})
		if err != nil {
			continue
		}
		job.RedisKey = keys[i]
		jobs = append(jobs, job)
	}

	if len(jobs) == 0 {
		return
	}

	hasLog := make([]*redis.IntCmd, len(jobs))
	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for i, job := range jobs {
			hasLog[i] = pipe.Exists(fmt.Sprintf("%s:jobs:%s:log", redisHeader, job.Id))
		}
		return nil
	})
	if err != nil {
		return
	}

	for i, result := range hasLog {
		jobs[i].HasLog = result.Val() > 0
	}

	return
}
