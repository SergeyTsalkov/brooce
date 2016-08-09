package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"brooce/config"

	redis "gopkg.in/redis.v3"
)

func (task *runnableTask) WriteLog(str string) {
	task.Write([]byte(str))
}

func (task *runnableTask) Write(p []byte) (lenP int, err error) {
	task.bufferLock.Lock()
	defer task.bufferLock.Unlock()

	lenP = len(p)
	//log.Printf("Task %v: %v", task.Id, string(p))

	if task.Id == "" {
		return
	}
	if config.Config.RedisOutputLog.DropDone && config.Config.RedisOutputLog.DropFailed {
		return
	}

	return task.buffer.Write(p)
}

func (task *runnableTask) Flush() {
	task.bufferLock.Lock()
	defer task.bufferLock.Unlock()

	if task.buffer.Len() == 0 {
		return
	}

	key := fmt.Sprintf("%s:jobs:%s:log", redisHeader, task.Id)

	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.Append(key, task.buffer.String())
		pipe.Expire(key, time.Duration(config.Config.RedisOutputLog.ExpireAfter)*time.Second)
		return nil
	})

	if err == nil {
		task.buffer.Reset()
	} else {
		log.Println("redis error:", err)
	}

	return
}

func (task *runnableTask) StartFlushingLog() {
	task.running = true
	task.buffer = &bytes.Buffer{}

	go func() {
		for task.running {
			time.Sleep(5 * time.Second)
			task.Flush()
		}
	}()
}

func (task *runnableTask) StopFlushingLog() {
	task.running = false
	task.Flush()
}

func (task *runnableTask) GenerateId() (err error) {
	var counter int64
	counter, err = redisClient.Incr(redisHeader + ":counter").Result()

	if err == nil {
		task.Id = fmt.Sprintf("%v", counter)
	}
	return
}
