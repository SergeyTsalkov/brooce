package web

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

func showlogHandler(req *http.Request) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}

	path := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(path) < 2 {
		err = fmt.Errorf("Invalid path")
		return
	}

	jobId := path[1]

	var lines []string
	lines, err = redisClient.LRange(fmt.Sprintf("%s:jobs:%s:log", redisHeader, jobId), 0, -1).Result()
	output := strings.TrimSpace(strings.Join(lines, ""))

	err = templates.ExecuteTemplate(buf, "showlog", output)
	return

}
