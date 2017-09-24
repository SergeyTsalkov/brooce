package runnabletask

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"brooce/config"

	redis "gopkg.in/redis.v5"
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

	if task.Id != "" && (!config.Config.RedisOutputLog.DropDone || !config.Config.RedisOutputLog.DropFailed) {
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
