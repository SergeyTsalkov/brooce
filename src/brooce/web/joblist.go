package web

import (
	"bytes"
	"net/http"

	"brooce/heartbeat"
)

type failedpageOutputType struct {
	Queues         map[string]*listQueueType
	RunningJobs    []*runningJobType
	RunningWorkers []*heartbeat.HeartbeatType
	TotalThreads   int
}

func failedpageHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}
	output := &mainpageOutputType{}
	output.Queues, err = listQueues()
	if err != nil {
		return
	}
	output.RunningJobs, err = listRunningJobs()
	if err != nil {
		return
	}
	output.RunningWorkers, err = listRunningWorkers()
	if err != nil {
		return
	}

	for _, worker := range output.RunningWorkers {
		output.TotalThreads += worker.TotalThreads()
	}

	err = templates.ExecuteTemplate(buf, "failedpage", output)
	return
}
