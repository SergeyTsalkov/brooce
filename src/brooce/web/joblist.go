package web

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"brooce/task"
)

type joblistOutputType struct {
	ListType  string
	QueueName string
	Page      int
	Pages     int
	Length    int

	Jobs []*task.Task
}

func joblistHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}

	path := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(path) < 2 {
		err = fmt.Errorf("Invalid path")
		return
	}

	listType := path[0]
	queueName := path[1]

	output := &joblistOutputType{
		QueueName: queueName,
		ListType:  listType,
	}
	err = output.listJobs(1)
	if err != nil {
		return
	}

	err = templates.ExecuteTemplate(buf, "joblist", output)
	return
}

func (output *joblistOutputType) listJobs(page int) (err error) {
	perPage := 100

	var values []string
	redisKey := fmt.Sprintf("%s:queue:%s:%s", redisHeader, output.QueueName, output.ListType)
	values, err = redisClient.LRange(redisKey, int64((page-1)*perPage), int64(page*perPage-1)).Result()
	if err != nil {
		return
	}

	for _, value := range values {
		job, err := task.NewFromJson(value)
		if err != nil {
			continue
		}

		output.Jobs = append(output.Jobs, job)
	}

	return
}
