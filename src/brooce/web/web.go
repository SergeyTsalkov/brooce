package web

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	myredis "brooce/redis"
	"brooce/task"
	"brooce/web/tpl"

	redis "gopkg.in/redis.v3"
)

var redisClient = myredis.Get()
var redisHeader = "brooce"

var reqHandler = http.NewServeMux()
var templates = tpl.Get()

var serv = &http.Server{
	Addr:         ":8080",
	Handler:      reqHandler,
	ReadTimeout:  10 * time.Second,
	WriteTimeout: 10 * time.Second,
}

func Start() {
	reqHandler.HandleFunc("/", makeHandler(mainpageHandler))

	go func() {
		err := serv.ListenAndServe()
		if err != nil {
			log.Fatalln(err)
		}
	}()
}

func makeHandler(fn func(req *http.Request) (*bytes.Buffer, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf, err := fn(r)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.Copy(w, buf)
	}
}

type mainpageOutputType struct {
	ListQueues      map[string]*listQueueType
	ListRunningJobs []*runningJobType
}

func mainpageHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}
	output := &mainpageOutputType{}
	output.ListQueues, err = listQueues()
	if err != nil {
		return
	}
	output.ListRunningJobs, err = listRunningJobs()
	if err != nil {
		return
	}

	err = templates.ExecuteTemplate(buf, "mainpage", output)
	return
}

type listQueueType struct {
	QueueName     string
	Pending       int64
	Running       int
	pendingResult *redis.IntCmd
	runningResult *redis.StringSliceCmd
}

func listQueues() (list map[string]*listQueueType, err error) {
	list = map[string]*listQueueType{}
	var results []string
	results, err = redisClient.Keys(redisHeader + ":queue:*").Result()
	if err != nil {
		return
	}

	for _, result := range results {
		parts := strings.Split(result, ":")
		if len(parts) < 3 {
			continue
		}

		list[parts[2]] = &listQueueType{QueueName: parts[2]}
	}

	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for _, queue := range list {
			queue.pendingResult = pipe.LLen(fmt.Sprintf("%s:queue:%s:pending", redisHeader, queue.QueueName))
			queue.runningResult = pipe.Keys(fmt.Sprintf("%s:queue:%s:working:*", redisHeader, queue.QueueName))
		}
		return nil
	})
	if err != nil {
		return
	}

	for _, queue := range list {
		queue.Pending = queue.pendingResult.Val()
		queue.Running = len(queue.runningResult.Val())
	}

	return
}

type runningJobType struct {
	RedisKey   string
	WorkerName string
	QueueName  string
	Task       *task.Task
	task       *redis.StringCmd
}

func listRunningJobs() (jobs []*runningJobType, err error) {
	var results []string
	results, err = redisClient.Keys(redisHeader + ":queue:*:working:*").Result()
	if err != nil {
		return
	}

	for _, result := range results {
		parts := strings.Split(result, ":")
		if len(parts) < 5 {
			continue
		}

		job := &runningJobType{
			RedisKey:   result,
			WorkerName: parts[4],
			QueueName:  parts[2],
		}

		jobs = append(jobs, job)
	}

	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for _, job := range jobs {
			job.task = pipe.LIndex(job.RedisKey, 0)
		}
		return nil
	})
	if err != nil {
		return
	}

	for _, job := range jobs {
		job.Task, err = task.NewFromJson(job.task.Val())
		if err != nil {
			return
		}
	}

	return
}
