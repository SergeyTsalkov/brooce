package web

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"brooce/task"
)

type joblistOutputType struct {
	ListType  string
	QueueName string
	Page      int64
	Pages     int64
	PerPage   int64
	Length    int64
	Start     int64
	End       int64
	Query     string

	URL *url.URL

	Jobs []*task.Task
}

func joblistHandler(req *http.Request, rep *httpReply) (err error) {
	path := splitUrlPath(req.URL.Path)
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
		PerPage:   joblistPerPage(req),
		URL:       req.URL,
	}

	err = output.listJobs()
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

func joblistPerPage(req *http.Request) (perpage int64) {
	perpage = 10

	perpageCookie, err := req.Cookie("perpage")
	if err != nil {
		return
	}

	perpage, _ = strconv.ParseInt(perpageCookie.Value, 10, 0)
	if perpage < 1 || perpage > 100 {
		perpage = 10
	}

	return
}

func (output *joblistOutputType) LinkParamsForPage(page int64) template.URL {
	if output.URL == nil {
		return template.URL("")
	}

	q := output.URL.Query()
	q.Set("page", strconv.Itoa(int(page)))

	return template.URL(q.Encode())
}

func (output *joblistOutputType) LinkParamsForPrevPage(page int64) template.URL {
	return output.LinkParamsForPage(page - 1)
}

func (output *joblistOutputType) LinkParamsForNextPage(page int64) template.URL {
	return output.LinkParamsForPage(page + 1)
}

func (output *joblistOutputType) listJobs() (err error) {
	reverse := (output.ListType == "pending")
	redisKey := fmt.Sprintf("%s:queue:%s:%s", redisHeader, output.QueueName, output.ListType)

	output.Length, err = redisClient.LLen(redisKey).Result()
	if err != nil {
		return err
	}

	output.pageCalculate()

	if output.Length == 0 {
		return
	}

	rangeStart := output.Start - 1
	rangeEnd := output.End - 1

	if reverse {
		rangeStart, rangeEnd = (rangeEnd+1)*-1, (rangeStart+1)*-1
	}

	var jobs []string
	jobs, err = redisClient.LRange(redisKey, rangeStart, rangeEnd).Result()
	if err != nil {
		return err
	}

	output.Jobs = make([]*task.Task, len(jobs))

	for i, value := range jobs {
		job, err := task.NewFromJson(value, output.QueueName)
		if err != nil {
			continue
		}

		if reverse {
			output.Jobs[len(jobs)-i-1] = job
		} else {
			output.Jobs[i] = job
		}
	}

	task.PopulateHasLog(output.Jobs)
	return
}

func (output *joblistOutputType) pageCalculate() {
	if output.Page < 1 {
		output.Page = 1
	}
	if output.PerPage < 1 {
		output.PerPage = 1
	}

	output.Pages = int64(math.Ceil(float64(output.Length) / float64(output.PerPage)))

	if output.Length == 0 {
		output.Start = 0
		output.End = 0
		output.Page = 0
		return
	}

	for {
		output.Start = (output.Page-1)*output.PerPage + 1
		output.End = output.Page * output.PerPage

		if output.Start > output.Length {
			output.Page = output.Pages
			continue
		}

		break
	}

	if output.End > output.Length {
		output.End = output.Length
	}
}
