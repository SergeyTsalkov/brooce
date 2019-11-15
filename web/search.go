package web

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	myredis "brooce/redis"
	"brooce/task"
)

func searchHandler(req *http.Request, rep *httpReply) (err error) {
	query, queueName, listType, page := searchQueryParams(req.URL.RawQuery)

	output := &joblistOutputType{
		QueueName: queueName,
		ListType:  listType,
		Query:     query,
		Page:      page,
		PerPage:   joblistPerPage(req),
		URL:       req.URL,
	}

	output.searchQueueAndPopulateResults()

	err = templates.ExecuteTemplate(rep, "joblist", output)
	return
}

func (output *joblistOutputType) searchQueueAndPopulateResults() (err error) {
	jsonResults := searchQueue(output.Query, output.QueueName, output.ListType)

	output.Length = int64(len(jsonResults))
	output.pageCalculate()

	if output.Length == 0 {
		return
	}

	start := output.Start - 1
	end := output.End

	output.Jobs = []*task.Task{}
	for _, json := range jsonResults[start:end] {
		t, err := task.NewFromJson(json, output.QueueName)
		if err != nil {
			log.Printf("Couldn't construct task.Task from %+v", json)
			continue
		}

		output.Jobs = append(output.Jobs, t)
	}

	task.PopulateHasLog(output.Jobs)
	return
}

func searchQueryParams(rq string) (query string, queue string, listType string, page int64) {
	params, err := url.ParseQuery(rq)
	if err != nil {
		log.Printf("Malformed URL query: %s err: %s", rq, err)
		return "", "", "done", 1
	}

	query = params.Get("q")
	queue = params.Get("queue")
	listType = params.Get("listType")
	if listType == "" {
		listType = "done"
	}

	page, _ = strconv.ParseInt(params.Get("page"), 10, 0)
	if page < 1 {
		page = 1
	}

	return query, queue, listType, page
}

func searchQueue(query, queueName, listType string) []string {
	r := myredis.Get()
	queueKey := fmt.Sprintf("%s:queue:%s:%s", redisHeader, queueName, listType)

	found := []string{}
	vals := r.LRange(queueKey, 0, -1).Val()

	for _, v := range vals {
		if strings.Contains(v, query) {
			found = append(found, v)
		}
	}

	return found
}
