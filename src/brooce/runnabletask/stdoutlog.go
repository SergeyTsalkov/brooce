package runnabletask

import (
	"fmt"
	"strings"
	"time"
)

type runnableTaskStdoutLog struct {
	*RunnableTask
	firstLineDone bool
}

func (task *runnableTaskStdoutLog) Write(p []byte) (lenP int, err error) {
	lenP = len(p)
	str := string(p)

	prefix := fmt.Sprintf("%s> ", time.Now().Format(tsFormat))
	str = strings.Replace(str, "\n", "\n"+prefix, -1)

	if !task.firstLineDone {
		str = prefix + str
		task.firstLineDone = true
	}

	_, err = task.RunnableTask.Write([]byte(str))
	return
}
