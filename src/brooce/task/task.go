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
	Id       string   `json:"id,omitempty"`
	Command  string   `json:"command"`
	Timeout  int64    `json:"timeout,omitempty"`
	MaxTries int      `json:"max_tries"`
	Locks    []string `json:"locks,omitempty"`

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
	if err == nil {
		return task, nil
	}

	if words := strings.Fields(str); len(words) == 0 {
		return task, fmt.Errorf("Invalid task: %s", str)
	}

	task.Command = str

	if task.Timeout == 0 {
		task.Timeout = int64(defaultOpts.Timeout)
		// task.Timeout = config.Config.GlobalJobOptions.Timeout
	}

	if task.MaxTries == 0 {
		task.MaxTries = defaultOpts.MaxTries
	}

	return task, nil
}

func (task *Task) TimeoutSeconds() time.Duration {
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
