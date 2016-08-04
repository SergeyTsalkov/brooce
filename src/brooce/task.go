package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"brooce/config"
	tasklib "brooce/task"

	redis "gopkg.in/redis.v3"
)

type runnableTask struct {
	*tasklib.Task
	workingList string
}

func (task *runnableTask) Run() (exitCode int, err error) {
	if len(task.Command) == 0 {
		return
	}

	if task.Id == "" {
		err = task.GenerateId()
		if err != nil {
			return
		}
	}

	name := task.Command[0]
	args := []string{}

	if len(task.Command) > 1 {
		args = task.Command[1:]
	}

	starttime := time.Now()
	task.StartTime = starttime.Unix()
	err = redisClient.LSet(task.workingList, 0, task.Json()).Err()
	if err != nil {
		return
	}

	log.Printf("Starting task %v: %v", task.Id, task.FullCommand())
	defer func() {
		log.Printf("Task %v exited after %v with exitcode %v", task.Id, time.Since(starttime), exitCode)
	}()

	cmd := exec.Command(name, args...)

	if task.Id != "" {
		cmd.Stdout = task
		cmd.Stderr = task
	}

	done := make(chan error)
	err = cmd.Start()
	if err != nil {
		return
	}

	go func() {
		done <- cmd.Wait()
	}()

	timeoutSeconds := task.Timeout
	if timeoutSeconds == 0 {
		timeoutSeconds = config.Config.Timeout
	}
	timeout := time.Duration(timeoutSeconds) * time.Second

	select {
	case err = <-done:
		//finished normally, do nothing!
	case <-time.After(timeout):
		//timed out!
		cmd.Process.Kill()
		err = fmt.Errorf("timeout after %v", timeout)
	}

	task.EndTime = time.Now().Unix()

	if msg, ok := err.(*exec.ExitError); ok {
		exitCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}

	return
}

func (task *runnableTask) Write(p []byte) (int, error) {
	log.Printf("Task %v: %v", task.Id, string(p))

	key := strings.Join([]string{redisHeader, "jobs", task.Id, "log"}, ":")

	redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.RPush(key, string(p))
		pipe.Expire(key, 7*24*time.Hour)
		return nil
	})

	return len(p), nil
}

func (task *runnableTask) GenerateId() (err error) {
	var counter int64
	counter, err = redisClient.Incr(redisHeader + ":counter").Result()

	if err == nil {
		task.Id = fmt.Sprintf("%v", counter)
	}
	return
}
