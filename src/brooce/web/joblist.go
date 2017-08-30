package web

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"brooce/config"
	"brooce/task"

	redis "gopkg.in/redis.v5"
)

type joblistOutputType struct {
	ListType  string
	QueueName string
	Page      int64
	Pages     int64
	Length    int64
	Start     int64
	End       int64
	Query     string

	URL *url.URL

	Jobs []*task.Task
}

func joblistHandler(req *http.Request, rep *httpReply) (err error) {
	path := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(path) < 2 {
		err = fmt.Errorf("Invalid path")
		return
	}

	listType := path[0]
	queueName := path[1]

	page := joblistQueryParams(req.URL.RawQuery)

	output := &joblistOutputType{
		QueueName: queueName,
		ListType:  listType,
		Page:      int64(page),
		URL:       req.URL,
	}
	err = output.listJobs(listType == "pending")
	if err != nil {
		return
	}

	err = templates.ExecuteTemplate(rep, "joblist", output)
	return
}

func joblistQueryParams(rq string) (page int) {
	params, err := url.ParseQuery(rq)
	if err != nil {
		log.Printf("Malformed URL query: %s err: %s", rq, err)
		return 1
	}

	page = 1
	if pg, err := strconv.Atoi(params.Get("page")); err == nil && pg > 1 {
		page = pg
	}

	return page
}

func (output *joblistOutputType) LinkParamsForPage(page int64) template.URL {
	if output.URL == nil {
		return template.URL("")
	}

	q := output.URL.Query()
	q.Set("page", strconv.Itoa(int(page)))

	return template.URL(q.Encode())
}

func (output *joblistOutputType) listJobs(reverse bool) (err error) {
	var perPage int64 = 10
	output.Start = (output.Page-1)*perPage + 1
	output.End = output.Page * perPage

	redisKey := fmt.Sprintf("%s:queue:%s:%s", redisHeader, output.QueueName, output.ListType)

	rangeStart := (output.Page - 1) * perPage
	rangeEnd := output.Page*perPage - 1

	if reverse {
		rangeStart, rangeEnd = (rangeEnd+1)*-1, (rangeStart+1)*-1
	}

	var lengthResult *redis.IntCmd
	var rangeResult *redis.StringSliceCmd
	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		lengthResult = pipe.LLen(redisKey)
		rangeResult = pipe.LRange(redisKey, rangeStart, rangeEnd)
		return nil
	})
	if err != nil {
		return
	}

	output.Length = lengthResult.Val()
	output.Pages = int64(math.Ceil(float64(output.Length) / float64(perPage)))
	if output.End > output.Length {
		output.End = output.Length
	}
	if output.Start > output.Length {
		output.Start = output.Length
	}

	rangeLength := len(rangeResult.Val())
	output.Jobs = make([]*task.Task, rangeLength)

	if len(output.Jobs) == 0 {
		return
	}

	for i, value := range rangeResult.Val() {
		opts := config.Config.LocalOptionsForQueue(output.QueueName)
		job, err := task.NewFromJson(value, opts)
		if err != nil {
			continue
		}

		if reverse {
			output.Jobs[rangeLength-i-1] = job
		} else {
			output.Jobs[i] = job
		}
	}

	hasLog := make([]*redis.BoolCmd, len(output.Jobs))
	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for i, job := range output.Jobs {
			hasLog[i] = pipe.Exists(fmt.Sprintf("%s:jobs:%s:log", redisHeader, job.Id))
		}
		return nil
	})
	if err != nil {
		return
	}

	for i, result := range hasLog {
		output.Jobs[i].HasLog = result.Val()
	}

	return
}
