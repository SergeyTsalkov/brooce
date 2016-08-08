package listing

import (
	"brooce/config"
	myredis "brooce/redis"
	"brooce/task"

	redis "gopkg.in/redis.v3"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

func RunningJobs() (jobs []*task.Task, err error) {
	jobs = []*task.Task{}

	var keys []string
	keys, err = redisClient.Keys(redisHeader + ":queue:*:working:*").Result()
	if err != nil {
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
		job, err := task.NewFromJson(value.Val())
		if err != nil {
			continue
		}
		job.RedisKey = keys[i]
		jobs = append(jobs, job)
	}

	return
}
