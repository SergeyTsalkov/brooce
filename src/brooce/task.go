package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"brooce/config"
	"brooce/lock"
	"brooce/prefixwriter"
	tasklib "brooce/task"

	redis "gopkg.in/redis.v3"
)

var tsFormat = "2006-01-02 15:04:05"

type runnableTask struct {
	*tasklib.Task
	workingList string
	threadName  string
	queueName   string
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

	starttime := time.Now()
	task.StartTime = starttime.Unix()
	err = redisClient.LSet(task.workingList, 0, task.Json()).Err()
	if err != nil {
		return
	}

	var gotLock bool
	gotLock, err = lock.GrabLocks(task.Locks)
	if err != nil {
		return
	}
	if !gotLock {
		exitCode = 75
		return
	}
	defer lock.ReleaseLocks(task.Locks)

	log.Printf("Starting task %v: %v", task.Id, task.Command)
	defer func() {
		finishtime := time.Now()
		runtime := finishtime.Sub(starttime)
		log.Printf("Task %v exited after %v with exitcode %v", task.Id, runtime, exitCode)

		task.WriteLog(fmt.Sprintf("\n*** EXITCODE:[%d] COMPLETED_AT:[%s] RUNTIME:[%s]\n",
			exitCode,
			finishtime.Format(tsFormat),
			runtime,
		))
	}()

	task.WriteLog(fmt.Sprintf("\n\n*** COMMAND:[%s] STARTED_AT:[%s] WORKER_THREAD:[%s] QUEUE:[%s]\n",
		task.Command,
		starttime.Format(tsFormat),
		task.threadName,
		task.queueName,
	))

	cmd := exec.Command("bash", "-c", task.Command)
	cmd.Stdout = &prefixwriter.PrefixWriter{Writer: task, Prefix: "--ts--> ", TsFormat: tsFormat}
	cmd.Stderr = cmd.Stdout

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

func (task *runnableTask) WriteLog(str string) {
	task.Write([]byte(str))
}

func (task *runnableTask) Write(p []byte) (int, error) {
	//log.Printf("Task %v: %v", task.Id, string(p))

	if task.Id != "" {
		key := strings.Join([]string{redisHeader, "jobs", task.Id, "log"}, ":")

		redisClient.Pipelined(func(pipe *redis.Pipeline) error {
			pipe.RPush(key, string(p))
			pipe.Expire(key, 7*24*time.Hour)
			return nil
		})
	}

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
