package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"brooce/config"
	myredis "brooce/redis"
	tasklib "brooce/task"

	redis "gopkg.in/redis.v6"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName
var RedisKeyEnabled = redisHeader + ":cron:jobs"
var RedisKeyDisabled = redisHeader + ":cron:disabledjobs"

type CronType struct {
	Name string

	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string

	Queue       string
	Command     string
	Locks       []string
	Timeout     int64
	MaxTries    int
	KillOnDelay *bool
	Disabled    bool

	Raw string
}

func (cron *CronType) Disable() (err error) {
	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.HDel(RedisKeyEnabled, cron.Name)
		pipe.HSet(RedisKeyDisabled, cron.Name, cron.Raw)
		return nil
	})
	return
}

func (cron *CronType) Enable() (err error) {
	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.HDel(RedisKeyDisabled, cron.Name)
		pipe.HSet(RedisKeyEnabled, cron.Name, cron.Raw)
		return nil
	})
	return
}

func (cron *CronType) Delete() (err error) {
	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.HDel(RedisKeyDisabled, cron.Name)
		pipe.HDel(RedisKeyEnabled, cron.Name)
		return nil
	})
	return
}

func (cron *CronType) Run() (err error) {
	pendingList := fmt.Sprintf("%s:queue:%s:pending", redisHeader, cron.Queue)
	err = redisClient.LPush(pendingList, cron.Task().Json()).Err()
	return
}

func Get(name string) (cron *CronType, err error) {
	var line string
	disabled := false

	line, err = redisClient.HGet(RedisKeyEnabled, name).Result()
	if err != nil {
		line, err = redisClient.HGet(RedisKeyDisabled, name).Result()

		if err != nil {
			return
		}
		disabled = true
	}

	cron, err = ParseCronLine(name, line)
	if err != nil {
		return
	}

	cron.Disabled = disabled
	return
}

func ParseCronLine(name, line string) (*CronType, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("cron name can't be empty")
	}

	parts := strings.Fields(line)
	if len(parts) < 6 {
		return nil, fmt.Errorf("cron string seems invalid")
	}

	cron := &CronType{
		Name:       name,
		Raw:        line,
		Minute:     parts[0],
		Hour:       parts[1],
		DayOfMonth: parts[2],
		Month:      parts[3],
		DayOfWeek:  parts[4],
	}

	parts = parts[5:]

	for len(parts) > 0 && strings.Contains(parts[0], ":") {
		keyval := parts[0]
		parts = parts[1:]

		keyvalParts := strings.SplitN(keyval, ":", 2)
		key := keyvalParts[0]
		value := keyvalParts[1]

		switch strings.ToLower(key) {
		case "queue":
			cron.Queue = value
		case "locks":
			cron.Locks = strings.Split(value, ",")
		case "timeout":
			cron.Timeout, _ = strconv.ParseInt(value, 10, 64)
		case "maxtries":
			cron.MaxTries, _ = strconv.Atoi(value)
		case "killondelay":
			killondelay, _ := strconv.ParseBool(value)
			cron.KillOnDelay = &killondelay

		default:
			//nothing yet!
		}
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("cron string seems invalid")
	}
	if cron.Queue == "" {
		return nil, fmt.Errorf("cron without queue is invalid")
	}

	cron.Command = strings.Join(parts, " ")
	return cron, nil
}

func (cron *CronType) MatchTime(t time.Time) bool {
	t = t.UTC()

	if !cronTimeCompare(cron.Minute, t.Minute()) {
		return false
	}

	if !cronTimeCompare(cron.Hour, t.Hour()) {
		return false
	}

	if !cronTimeCompare(cron.DayOfMonth, t.Day()) {
		return false
	}

	if !cronTimeCompare(cron.Month, int(t.Month())) {
		return false
	}

	if !cronTimeCompare(cron.DayOfWeek, int(t.Weekday())) {
		return false
	}

	return true
}

func cronTimeCompare(cronstr string, timeval int) bool {
	if cronstr == "*" {
		return true
	}

	for _, cronval := range strings.Split(cronstr, ",") {
		if strings.Contains(cronval, "-") {
			cronValParts := strings.SplitN(cronval, "-", 2)
			start, _ := strconv.Atoi(cronValParts[0])
			end, _ := strconv.Atoi(cronValParts[1])
			if timeval >= start && timeval <= end {
				return true
			}
		} else if strings.HasPrefix(cronval, "*/") && len(cronval) > 2 {
			divisor, _ := strconv.Atoi(cronval[2:])
			if timeval%divisor == 0 {
				return true
			}
		} else {
			cronval, _ := strconv.Atoi(cronval)
			if cronval == timeval {
				return true
			}
		}
	}

	return false
}

func (cron *CronType) Task() *tasklib.Task {
	return &tasklib.Task{
		Command:     cron.Command,
		Cron:        cron.Name,
		Locks:       cron.Locks,
		Timeout:     cron.Timeout,
		MaxTries:    cron.MaxTries,
		KillOnDelay: cron.KillOnDelay,
	}
}
