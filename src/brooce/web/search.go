package web

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	myredis "brooce/redis"
	"brooce/task"
)

type PagedHits struct {
	Hits       []*task.Task
	Start      int
	End        int
	Pages      int
	PageSize   int
	PageWanted int
}

func searchHandler(req *http.Request, rep *httpReply) (err error) {
	query, queueName, listType, page := searchQueryParams(req.URL.RawQuery)

	hits := searchQueueForCommand(query, queueName, listType)
	pagedHits := newPagedHits(hits, 10, page)
	task.PopulateHasLog(pagedHits.Hits)

	output := &joblistOutputType{
		QueueName: queueName,
		ListType:  listType,
		Query:     query,
		Page:      int64(page),
		Pages:     int64(pagedHits.Pages),
		Jobs:      pagedHits.Hits,
		Start:     int64(pagedHits.Start),
		End:       int64(pagedHits.End),
		Length:    int64(len(hits)),

		URL: req.URL,
	}

	err = templates.ExecuteTemplate(rep, "joblist", output)
	return
}

func newPagedHits(hits []*task.Task, pageSize int, pageWanted int) *PagedHits {
	if pageWanted < 1 {
		pageWanted = 1
	}

	start := 1
	end := pageSize

	totalHits := len(hits)
	totalPages := int(math.Ceil(float64(totalHits) / float64(pageSize)))

	maxStart := (pageWanted - 1) * pageSize
	maxEnd := (pageWanted * pageSize) - 1

	if maxStart > totalHits {
		start = totalHits
	} else {
		start = maxStart
	}

	if (maxEnd + 1) > totalHits {
		end = totalHits
	} else {
		end = maxEnd + 1
	}

	// log.Printf("page %d: start: %d end: %d total pages: %d", pageWanted, start, end, totalPages)

	return &PagedHits{Hits: hits[start:end], Start: start + 1, End: end, PageWanted: pageWanted, Pages: totalPages}
}

func searchQueryParams(rq string) (query string, queue string, listType string, page int) {
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

	page = 1
	if pg, err := strconv.Atoi(params.Get("page")); err == nil && pg > 1 {
		page = pg
	}

	return query, queue, listType, page
}

func searchQueueForCommand(query, queueName, listType string) []*task.Task {
	r := myredis.Get()
	queueKey := fmt.Sprintf("%s:queue:%s:%s", redisHeader, queueName, listType)

	found := []*task.Task{}
	vals := r.LRange(queueKey, 0, -1).Val()

	for _, v := range vals {
		t, err := task.NewFromJson(v, queueName)

		if err != nil {
			log.Printf("Couldn't construct task.Task from %+v", v)
			continue
		}

		if strings.Contains(t.Command, query) {
			found = append(found, t)
		}
	}

	return found
}
