package task

import (
	"encoding/json"
	"log"
	"strings"
)

type Task struct {
	Id      string   `json:"id,omitempty"`
	Command []string `json:"command"`
	Timeout int      `json:"timeout,omitempty"`

	StartTime int64 `json:"start_time,omitempty"`
	EndTime   int64 `json:"end_time,omitempty"`

	Raw    string `json:"-"`
	HasLog bool   `json:"-"`
}

func NewFromJson(str string) (task *Task, err error) {
	task = &Task{}
	err = json.Unmarshal([]byte(str), task)
	if err != nil {
		return
	}
	task.Raw = str
	return
}

func (task *Task) Json() string {
	bytes, err := json.Marshal(task)
	if err != nil {
		log.Fatalln(err)
	}

	return string(bytes)
}

func (task *Task) FullCommand() string {
	if task.Command == nil {
		task.Command = []string{}
	}

	return strings.Join(task.Command, " ")
}
