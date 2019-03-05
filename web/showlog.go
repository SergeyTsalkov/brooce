package web

import (
	"fmt"
	"net/http"
)

func showlogHandler(req *http.Request, rep *httpReply) (err error) {
	path := splitUrlPath(req.URL.Path)
	if len(path) < 2 {
		err = fmt.Errorf("Invalid path")
		return
	}

	jobId := path[1]

	var output string
	output, err = redisClient.Get(fmt.Sprintf("%s:jobs:%s:log", redisLogHeader, jobId)).Result()
	if err != nil {
		return
	}

	err = templates.ExecuteTemplate(rep, "showlog", output)
	return
}
