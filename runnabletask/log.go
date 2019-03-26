package runnabletask

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

func (task *RunnableTask) WriteLog(str string) {
	//log.Printf("[%s] %s", task.WorkerThreadName(), strings.TrimSpace(str))
	task.Write([]byte(str))
}

func (task *RunnableTask) Write(p []byte) (lenP int, err error) {
	task.bufferLock.Lock()
	defer task.bufferLock.Unlock()
	lenP = len(p)

	if task.FileWriter != nil {
		task.FileWriter.Write(p)
	}

	if task.Id != "" && !task.NoRedisLog() {
		_, err = task.buffer.Write(p)
	}

	return
}

func (task *RunnableTask) Flush() {
	task.bufferLock.Lock()
	defer task.bufferLock.Unlock()

	if task.buffer.Len() == 0 {
		return
	}

	if task.LogKey() == "" {
		log.Println("Warning: trying to flush log but we have no task id!")
		task.buffer.Reset()
		return
	}

	_, err := redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.Append(task.LogKey(), task.buffer.String())

		if task.RedisLogExpireAfter() > 0 {
			pipe.Expire(task.LogKey(), time.Duration(task.RedisLogExpireAfter())*time.Second)
		}
		return nil
	})

	if err == nil {
		task.buffer.Reset()
	} else {
		log.Println("redis error:", err)
	}

	return
}

func (task *RunnableTask) StartFlushingLog() {
	task.running = true
	task.buffer = &bytes.Buffer{}

	go func() {
		for task.running {
			time.Sleep(5 * time.Second)
			task.Flush()
		}
	}()
}

func (task *RunnableTask) StopFlushingLog() {
	task.running = false
	task.Flush()
}

func (task *RunnableTask) GenerateId() (err error) {
	var counter int64
	counter, err = redisClient.Incr(redisHeader + ":counter").Result()

	if err == nil {
		task.Id = fmt.Sprintf("%v", counter)
	}
	return
}
