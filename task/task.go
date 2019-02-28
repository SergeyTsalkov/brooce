package task

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"brooce/config"
	myredis "brooce/redis"

	"github.com/go-redis/redis"
)

type Task struct {
	// user-settable config items
	Id      string   `json:"id,omitempty"`
	Command string   `json:"command"`
	Locks   []string `json:"locks,omitempty"`

	// user-settable job options that inherit from global
	config.JobOptions

	// machine-settable
	Tried     int    `json:"tried"`
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
	Cron      string `json:"cron,omitempty"`

	Raw      string `json:"-"`
	RedisKey string `json:"-"`
	HasLog   bool   `json:"-"`
}

func NewFromJson(str string, queueName string) (*Task, error) {
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

	task.JobOptions.Merge(config.Config.JobOptionsForQueue(queueName))
	task.JobOptions.Merge(config.Config.GlobalJobOptions)
	task.JobOptions.Merge(config.DefaultJobOptions)

	return task, nil
}

func PopulateHasLog(tasks []*Task) {
	tasksWithId := []*Task{}
	for _, task := range tasks {
		if task.LogKey() != "" {
			tasksWithId = append(tasksWithId, task)
		}
	}

	if len(tasksWithId) == 0 {
		return
	}

	hasLog := make([]*redis.IntCmd, len(tasksWithId))
	_, err := myredis.Get().Pipelined(func(pipe redis.Pipeliner) error {
		for i, task := range tasksWithId {
			hasLog[i] = pipe.Exists(task.LogKey())
		}
		return nil
	})
	if err != nil {
		log.Println("Warning: redis error when trying to check task logs:", err)
		return
	}

	for i, result := range hasLog {
		if result != nil {
			tasksWithId[i].HasLog = result.Val() > 0
		}
	}
}

func (task *Task) Json() string {
	bytes, err := json.Marshal(task)
	if err != nil {
		log.Fatalln(err)
	}

	return string(bytes)
}

func QueueNameFromRedisKey(redisKey string) string {
	parts := strings.SplitN(redisKey, ":", 5)
	if len(parts) < 5 {
		return ""
	}
	return parts[2]
}

func (task *Task) QueueName() string {
	return QueueNameFromRedisKey(task.RedisKey)
}

func (task *Task) WorkerThreadName() string {
	parts := strings.SplitN(task.RedisKey, ":", 5)
	if len(parts) < 5 {
		return ""
	}
	return parts[4]
}

func (task *Task) LogKey() string {
	if task.Id == "" {
		return ""
	}

	return fmt.Sprintf("%s:jobs:%s:log", config.Config.ClusterLogName, task.Id)
}
