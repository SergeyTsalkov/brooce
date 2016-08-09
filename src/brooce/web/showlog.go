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

	var output string
	output, err = redisClient.Get(fmt.Sprintf("%s:jobs:%s:log", redisHeader, jobId)).Result()
	if err != nil {
		return
	}

	err = templates.ExecuteTemplate(buf, "showlog", output)
	return
}
