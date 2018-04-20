package config

import (
	"fmt"
	"strings"
)

type ThreadType struct {
	Id      int        `json:"id"`
	Name    string     `json:"name"`
	Queue   string     `json:"queue"`
	Options JobOptions `json:"-"`
}

func (t *ThreadType) PendingList() string {
	return fmt.Sprintf("%s:queue:%s:pending", Config.ClusterName, t.Queue)
}

func (t *ThreadType) WorkingList() string {
	return fmt.Sprintf("%s:queue:%s:working:%s", Config.ClusterName, t.Queue, t.Name)
}

func (t *ThreadType) DoneList() string {
	return fmt.Sprintf("%s:queue:%s:done", Config.ClusterName, t.Queue)
}

func (t *ThreadType) FailedList() string {
	return fmt.Sprintf("%s:queue:%s:failed", Config.ClusterName, t.Queue)
}

func (t *ThreadType) DelayedList() string {
	return fmt.Sprintf("%s:queue:%s:delayed", Config.ClusterName, t.Queue)
}

var Threads = []ThreadType{}
var ThreadString string

func init_threads() {
	strQueueList := []string{}

	for _, q := range Config.Queues {
		for i := 0; i < q.Workers; i++ {

			thread := ThreadType{
				Id:      i,
				Name:    fmt.Sprintf("%v-%v-%v", Config.ProcName, q.Name, i),
				Queue:   q.Name,
				Options: Config.LocalOptionsForQueue(q.Name),
			}

			Threads = append(Threads, thread)
		}

		strQueueList = append(strQueueList, fmt.Sprintf("%v (x%v)", q.Name, q.Workers))
	}

	ThreadString = strings.Join(strQueueList, ", ")
}
