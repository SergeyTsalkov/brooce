package cronsched

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"brooce/config"
	"brooce/cron"
	"brooce/listing"
	myredis "brooce/redis"
	"brooce/util"

	redis "gopkg.in/redis.v5"
)

var redisHeader = config.Config.ClusterName
var redisClient = myredis.Get()

func Start() {
	go func() {
		for {
			util.SleepUntilNextMinute()
			err := scheduleCrons()
			if err != nil {
				log.Println("Cron scheduling error:", err)
			}
		}
	}()
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
		pipe.SetNX(lockKey, config.Config.ProcName, lockttl)
		lockValueCmd = pipe.Get(lockKey)
		return nil
	})
	if err != nil {
		return err
	}

	lockValue, _ := lockValueCmd.Result()

	if lockValue != config.Config.ProcName {
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
		scheduleCronsForTimeRange(pipe, start, end)
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

func scheduleCronsForTimeRange(pipe *redis.Pipeline, start time.Time, end time.Time) {
	crons, err := listing.Crons()
	if err != nil {
		log.Println("redis error:", err)
		return
	}

	toSchedule := map[string]*cron.CronType{}

	for t := start; !t.After(end); t = t.Add(time.Minute) {
		for cronName, cron := range crons {
			if cron.MatchTime(t) {
				toSchedule[cronName] = cron
			}
		}
	}

	if !start.Equal(end) && len(toSchedule) > 0 {
		log.Println("Cron is catching up! Scheduling jobs for the period from", start, "to", end)
	}

	for cronName, cronJob := range toSchedule {
		log.Printf("Scheduling job %s", cronName)

		pendingList := strings.Join([]string{redisHeader, "queue", cronJob.Queue, "pending"}, ":")
		pipe.LPush(pendingList, cronJob.Task().Json())
	}
}

func CronIsRunning(c *cron.CronType) (bool, error) {
	if c.Name == "" {
		return false, fmt.Errorf("cron has no name, can't determine if it's running")
	}

	jobs, err := listing.RunningJobs()
	if err != nil {
		return false, err
	}

	for _, job := range jobs {
		if c.Name == job.Cron {
			return true, nil
		}
	}

	return false, nil
}
