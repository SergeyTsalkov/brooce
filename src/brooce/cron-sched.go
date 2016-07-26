package main

import (
	"strconv"
	"strings"
	"time"

	redis "gopkg.in/redis.v3"
)

func cronner() {
	for {
		sleepUntil00()
		err := scheduleCrons()
		if err != nil {
			logger.Println("Cron scheduling error:", err)
		}
	}
}

func scheduleCrons() error {
	schedThroughKey := redisHeader + ":cron:scheduled_through"
	lockKey := redisHeader + ":cron:lock"
	lockttl := 90 * time.Second

	// if it's been longer than this since the scheduler last ran
	// then just start from here
	maxSchedCatchup := 24 * time.Hour

	var lockValueCmd *redis.StringCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.SetNX(lockKey, myProcName, lockttl)
		lockValueCmd = pipe.Get(lockKey)
		return nil
	})

	if err != nil {
		return err
	}

	lockValue, _ := lockValueCmd.Result()

	if lockValue != myProcName {
		return nil
	}

	schedThroughUnixtimeStr, _ := redisClient.Get(schedThroughKey).Result()
	schedThroughUnixtime, _ := strconv.ParseInt(schedThroughUnixtimeStr, 10, 64)

	start := time.Unix(schedThroughUnixtime, 0)
	start = zeroOutSeconds(start)
	start = start.Add(time.Minute)

	end := time.Now()
	end = zeroOutSeconds(end)

	if end.Sub(start) > maxSchedCatchup || start.After(end) {
		start = end
	}

	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		scheduleCronsForTimeRange(pipe, listActiveCrons(), start, end)
		pipe.Set(schedThroughKey, end.Unix(), maxSchedCatchup)
		pipe.Expire(lockKey, lockttl)
		return nil
	})

	return err
}

func zeroOutSeconds(t time.Time) time.Time {
	if t.Second() != 0 {
		t = t.Add(time.Duration(-1*t.Second()) * time.Second)
	}

	if t.Nanosecond() != 0 {
		t = t.Add(time.Duration(-1*t.Nanosecond()) * time.Nanosecond)
	}
	return t
}

func listActiveCrons() map[string]*cronType {
	crons := map[string]*cronType{}

	cronKeySet := redisHeader + ":cron:jobs:"
	cronKeys, err := redisClient.Keys(cronKeySet + "*").Result()
	if err != nil {
		return crons
	}

	for _, cronKey := range cronKeys {
		cronName := cronKey
		if strings.HasPrefix(cronName, cronKeySet) {
			cronName = strings.Replace(cronName, cronKeySet, "", 1)
		}

		cronLine, err := redisClient.Get(cronKey).Result()
		if err != nil {
			continue
		}

		cron, _ := parseCronLine(cronLine)
		if cron != nil {
			crons[cronName] = cron
		}
	}

	return crons
}

func scheduleCronsForTimeRange(pipe *redis.Pipeline, crons map[string]*cronType, start time.Time, end time.Time) {
	toSchedule := map[string]*cronType{}

	for t := start; !t.After(end); t = t.Add(time.Minute) {
		if !start.Equal(end) {
			logger.Println("Scheduling cron for", t)
		}

		for cronName, cron := range crons {
			if cron.matchTime(t) {
				toSchedule[cronName] = cron
			}
		}
	}

	for cronName, cron := range toSchedule {
		logger.Println("Scheduling job", cronName, ":", strings.Join(cron.command, " "))

		if cron.queue == "" {
			continue
		}

		pendingList := strings.Join([]string{redisHeader, "queue", cron.queue, "pending"}, ":")
		pipe.LPush(pendingList, cron.task().Json())
	}
}
