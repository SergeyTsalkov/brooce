package listing

import (
	"brooce/cron"
	"log"
)

func Crons() (map[string]*cron.CronType, error) {
	return crons(false)
}

func DisabledCrons() (map[string]*cron.CronType, error) {
	return crons(true)
}

func crons(disabled bool) (crons map[string]*cron.CronType, err error) {
	crons = map[string]*cron.CronType{}

	cronKey := cron.RedisKeyEnabled
	if disabled {
		cronKey = cron.RedisKeyDisabled
	}

	var results map[string]string
	results, err = redisClient.HGetAll(cronKey).Result()
	if err != nil || len(results) == 0 {
		return
	}

	for cronName, cronValue := range results {
		cron, err := cron.ParseCronLine(cronName, cronValue)
		if err != nil || cron == nil {
			log.Println("Warning: Unable to parse this cron job:", cronValue)
			continue
		}

		cron.Disabled = disabled
		crons[cronName] = cron
	}

	return
}
