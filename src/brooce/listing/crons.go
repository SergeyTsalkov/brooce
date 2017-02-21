package listing

import (
	"strings"

	"brooce/cron"

	redis "gopkg.in/redis.v5"
)

func Crons() (map[string]*cron.CronType, error) {
	return crons(false)
}

func DisabledCrons() (map[string]*cron.CronType, error) {
	return crons(true)
}

func crons(disabled bool) (crons map[string]*cron.CronType, err error) {
	crons = map[string]*cron.CronType{}

	var keys []string
	cronKeyPrefix := redisHeader + ":cron:jobs:"
	if disabled {
		cronKeyPrefix = strings.Replace(cronKeyPrefix, ":jobs:", ":disabledjobs:", 1)
	}

	keys, err = redisClient.Keys(cronKeyPrefix + "*").Result()
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

		if disabled {
			crons[cronName].Disabled = true
		}
	}

	return
}
