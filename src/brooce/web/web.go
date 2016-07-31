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
	ListQueues map[string]int64
}

func mainpageHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}
	output := &mainpageOutputType{}
	output.ListQueues, err = listQueues()
	if err != nil {
		return
	}

	err = templates.ExecuteTemplate(buf, "mainpage", output)
	return
}

func listQueues() (list map[string]int64, err error) {
	var results []string
	results, err = redisClient.Keys(redisHeader + ":queue:*").Result()
	if err != nil {
		return
	}

	var queues []string
	for _, result := range results {
		parts := strings.Split(result, ":")
		if len(parts) < 3 {
			continue
		}

		queues = append(queues, parts[2])
	}

	var lenResults []*redis.IntCmd
	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for _, queueName := range queues {
			result := pipe.LLen(fmt.Sprintf("%s:queue:%s:pending", redisHeader, queueName))
			lenResults = append(lenResults, result)
		}
		return nil
	})
	if err != nil {
		return
	}

	list = map[string]int64{}
	for i, queueName := range queues {
		list[queueName] = lenResults[i].Val()
	}

	return
}
