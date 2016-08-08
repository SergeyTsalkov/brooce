package listing

import (
	"strings"

	"brooce/cron"

	redis "gopkg.in/redis.v3"
)

func Crons() (crons map[string]*cron.CronType) {
	crons = map[string]*cron.CronType{}

	cronKeyPrefix := redisHeader + ":cron:jobs:"
	keys, err := redisClient.Keys(cronKeyPrefix + "*").Result()
	if err != nil {
		return
	}

	cronValues := make([]*redis.StringCmd, len(keys))
	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for i, key := range keys {
			cronValues[i] = pipe.Get(key)
		}
		return nil
	})
	if err != nil {
		return
	}

	for i, key := range keys {
		cronName := strings.TrimPrefix(key, cronKeyPrefix)
		cronValue := cronValues[i].Val()

		cron, err := cron.ParseCronLine(cronName, cronValue)
		if err == nil && cron != nil {
			crons[cronName] = cron
		}
	}

	return
}
