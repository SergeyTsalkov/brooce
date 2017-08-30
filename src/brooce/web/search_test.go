package web

import (
	"fmt"
	"log"
	"testing"

	"brooce/config"
	"brooce/task"
)

// These tests imply a queue named brooce:queue:common:done containing at least one job whose
// command includes the word fart ie:
// redis-cli lpush '{"id":"54","command":"sleep 3 && echo fart","timeout":10,"max_tries":2,"tried":1,"start_time":1504040030,"end_time":1504040033}'

func TestSearchQueueForCommand(t *testing.T) {
	tasks := searchQueueForCommand("brooce:queue:common:done", "fart")
	t.Logf("TASKS: %+v", tasks)
	if len(tasks) < 1 {
		t.Errorf("This was supposed to contain at least one task that matched 'fart': %+v", tasks)
	}
}

func TestSearchForCommand(t *testing.T) {
	results := searchForCommand("config-server")
	t.Logf("Found %d hits", results.Count())
	t.Logf("First hit: %+v", results.Queues)
	t.Logf("Total entries searched: %d", results.TotalSearched)
	if results.Count() < 1 {
		t.Errorf("I FOUND NONE DUMMY")
	}
}

func TestPageinateHits(t *testing.T) {
	hits := make([]*task.Task, 106)
	for i := 0; i < 106; i++ {
		tt, _ := task.NewFromJson(fmt.Sprintf("echo 'hello %d'", i), config.JobOptions{})
		hits[i] = tt
	}

	var ph *PagedHits
	for i := 0; i < 7; i++ {
		ph = paginateHits(hits, 25, i)
		log.Printf("%d: start: %d end: %d page: %d pages: %d", i, ph.Start, ph.End, i, ph.Pages)
	}
}
