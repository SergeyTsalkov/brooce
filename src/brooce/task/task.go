package task

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Task struct {
	Id      string   `json:"id,omitempty"`
	Command string   `json:"command"`
	Timeout int64    `json:"timeout,omitempty"`
	Locks   []string `json:"locks,omitempty"`

	Cron      string `json:"cron,omitempty"`
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`

	Raw      string `json:"-"`
	RedisKey string `json:"-"`
	HasLog   bool   `json:"-"`
}

func NewFromJson(str string) (*Task, error) {
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
	return task, nil
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
