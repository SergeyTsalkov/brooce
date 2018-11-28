package runnabletask

import (
	"fmt"
	"strings"
	"time"
)

type runnableTaskStdoutLog struct {
	*RunnableTask
}

func (task *runnableTaskStdoutLog) Write(p []byte) (lenP int, err error) {
	lenP = len(p)
	str := string(p)

	prefix := fmt.Sprintf("%s> ", time.Now().Format(tsFormat))
	str = strings.TrimSpace(str)
	str = strings.Replace(str, "\n", "\n"+prefix, -1)
	str = prefix + str + "\n"

	_, err = task.RunnableTask.Write([]byte(str))
	return
}
