package runnabletask

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"brooce/config"
	"brooce/lock"
	myredis "brooce/redis"
	tasklib "brooce/task"
)

var tsFormat = "2006-01-02 15:04:05"

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

type RunnableTask struct {
	*tasklib.Task
	FileWriter io.Writer

	buffer     *bytes.Buffer
	bufferLock sync.Mutex
	running    bool
}

func (task *RunnableTask) Run() (exitCode int, err error) {
	if len(task.Command) == 0 {
		return
	}

	if task.Id == "" {
		err = task.GenerateId()
		if err != nil {
			err = fmt.Errorf("Error in task.GenerateId: %v", err)
			return
		}
	}

	starttime := time.Now()
	task.StartTime = starttime.Unix()
	err = redisClient.LSet(task.RedisKey, 0, task.Json()).Err()
	if err != nil {
		err = fmt.Errorf("Error updating working key 0: %v", err)
		return
	}

	var gotLock bool
	gotLock, err = lock.GrabLocks(task.Locks)
	if err != nil {
		err = fmt.Errorf("Error grabbing locks: %v", err)
		return
	}
	if !gotLock {
		exitCode = 75
		return
	}
	defer lock.ReleaseLocks(task.Locks)

	task.StartFlushingLog()
	defer func() {
		finishtime := time.Now()
		runtime := finishtime.Sub(starttime)

		task.WriteLog(fmt.Sprintf("*** COMPLETED_AT:[%s] RUNTIME:[%s] EXITCODE:[%d] ERROR:[%v]\n",
			finishtime.Format(tsFormat),
			runtime,
			exitCode,
			err,
		))
		task.StopFlushingLog()
	}()

	task.WriteLog(fmt.Sprintf("\n\n*** COMMAND:[%s] STARTED_AT:[%s] WORKER_THREAD:[%s] QUEUE:[%s]\n",
		task.Command,
		starttime.Format(tsFormat),
		task.WorkerThreadName(),
		task.QueueName(),
	))

	cmd := exec.Command("bash", "-c", task.Command)
	cmd.Stdout = &runnableTaskStdoutLog{RunnableTask: task}
	cmd.Stderr = cmd.Stdout

	done := make(chan error)
	err = cmd.Start()
	if err != nil {
		return
	}

	go func() {
		done <- cmd.Wait()
	}()

	// timeoutSeconds := task.Timeout
	// if timeoutSeconds == 0 {

	// 	timeoutSeconds = int64(config.Config.GlobalJobOptions.Timeout)
	// }
	// timeout := time.Duration(timeoutSeconds) * time.Second
	timeout := task.TimeoutSeconds()

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
