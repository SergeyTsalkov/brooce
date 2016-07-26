package main

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	tasklib "brooce/task"

	redis "gopkg.in/redis.v3"
)

type runnableTask struct {
	*tasklib.Task
}

func (task *runnableTask) Run() (exitCode int) {
	if len(task.Command) == 0 {
		return
	}

	if task.Id == "" {
		counter, err := redisClient.Incr(redisHeader + ":counter").Result()

		if err != nil {
			logger.Printf("Unable to pick id for task, redis said: %v", err)
			exitCode = 1
			return
		}

		task.Id = fmt.Sprintf("%v", counter)
	}

	name := task.Command[0]
	args := []string{}

	if len(task.Command) > 1 {
		args = task.Command[1:]
	}

	logger.Printf("Starting task %v: %v", task.Id, task.FullCommand())
	starttime := time.Now()

	cmd := exec.Command(name, args...)

	if task.Id != "" {
		cmd.Stdout = task
		cmd.Stderr = task
	}

	err := cmd.Run()

	// Grab unix process exit code
	// WINDOWS BARFS HERE OH GOD
	if err != nil {
		if msg, ok := err.(*exec.ExitError); ok {
			exitCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
		}
	}

	logger.Printf("Task %v exited after %v with exitcode %v", task.Id, time.Since(starttime), exitCode)
	return
}

func (task *runnableTask) Write(p []byte) (int, error) {
	logger.Printf("Task %v: %v", task.Id, string(p))

	key := strings.Join([]string{redisHeader, "jobs", task.Id, "log"}, ":")

	redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.RPush(key, string(p))
		pipe.Expire(key, 7*24*time.Hour)
		return nil
	})

	return len(p), nil
}
