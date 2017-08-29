package listing

import (
	"brooce/config"
	myredis "brooce/redis"
	"brooce/task"
	"fmt"

	redis "gopkg.in/redis.v5"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

func RunningJobs() (jobs []*task.Task, err error) {
	jobs = []*task.Task{}

	var keys []string
	keys, err = redisClient.Keys(redisHeader + ":queue:*:working:*").Result()
	if err != nil || len(keys) == 0 {
		return
	}

	values := make([]*redis.StringCmd, len(keys))
	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for i, key := range keys {
			values[i] = pipe.LIndex(key, 0)
		}
		return nil
	})
	if err != nil {
		return
	}

	for i, value := range values {
		job, err := task.NewFromJson(value.Val(), config.JobOptions{})
		if err != nil {
			continue
		}
		job.RedisKey = keys[i]
		jobs = append(jobs, job)
	}

	hasLog := make([]*redis.BoolCmd, len(jobs))
	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for i, job := range jobs {
			hasLog[i] = pipe.Exists(fmt.Sprintf("%s:jobs:%s:log", redisHeader, job.Id))
		}
		return nil
	})
	if err != nil {
		return
	}

	for i, result := range hasLog {
		jobs[i].HasLog = result.Val()
	}

	return
}
