package task

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"brooce/config"
)

type Task struct {
	Id          string   `json:"id,omitempty"`
	Command     string   `json:"command"`
	Timeout     int64    `json:"timeout,omitempty"`
	MaxTries    int      `json:"max_tries"`
	Tried       int      `json:"tried"`
	Locks       []string `json:"locks,omitempty"`
	KillOnDelay *bool    `json:"killondelay,omitempty"`

	Cron      string `json:"cron,omitempty"`
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`

	Raw      string `json:"-"`
	RedisKey string `json:"-"`
	HasLog   bool   `json:"-"`
}

func NewFromJson(str string, defaultOpts config.JobOptions) (*Task, error) {
	task := &Task{}
	defer func() {
		task.Raw = str
	}()

	err := json.Unmarshal([]byte(str), task)
	if err != nil {
		// we have a non-JSON (plain-string) command.
		if words := strings.Fields(str); len(words) == 0 {
			return task, fmt.Errorf("Invalid task: %s", str)
		}

		task.Command = str
	}

	if task.Timeout == 0 {
		task.Timeout = int64(defaultOpts.Timeout)
	}

	if task.MaxTries == 0 {
		task.MaxTries = defaultOpts.MaxTries
	}

	if task.KillOnDelay == nil {
		task.KillOnDelay = defaultOpts.KillOnDelay
	}

	return task, nil
}

func (task *Task) TimeoutDuration() time.Duration {
	return time.Duration(task.Timeout) * time.Second
}

func (task *Task) Json() string {
	bytes, err := json.Marshal(task)
	if err != nil {
		log.Fatalln(err)
	}

	return string(bytes)
}

func (task *Task) QueueName() string {
	parts := strings.SplitN(task.RedisKey, ":", 5)
	if len(parts) < 5 {
		return ""
	}
	return parts[2]
}

func (task *Task) WorkerThreadName() string {
	parts := strings.SplitN(task.RedisKey, ":", 5)
	if len(parts) < 5 {
		return ""
	}
	return parts[4]
}
