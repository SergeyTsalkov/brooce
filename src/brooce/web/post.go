package web

import (
	"fmt"
	"net/http"
	"strings"

	myredis "brooce/redis"
)

func retryHandler(req *http.Request, rep *httpReply) (err error) {
	path := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(path) < 3 {
		err = fmt.Errorf("Invalid path")
		return
	}

	listType := path[1]
	queueName := path[2]

	removeKey := fmt.Sprintf("%s:queue:%s:%s", redisHeader, queueName, listType)
	pushKey := fmt.Sprintf("%s:queue:%s:pending", redisHeader, queueName)

	if item := req.FormValue("item"); item != "" {
		var count int64
		count, err = redisClient.LRem(removeKey, 1, item).Result()
		if err != nil {
			return
		}
		if count == 1 {
			redisClient.LPush(pushKey, item)
		}
	}

	return
}

func deleteHandler(req *http.Request, rep *httpReply) (err error) {
	path := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(path) < 3 {
		err = fmt.Errorf("Invalid path")
		return
	}

	listType := path[1]
	queueName := path[2]

	removeKey := fmt.Sprintf("%s:queue:%s:%s", redisHeader, queueName, listType)

	if item := req.FormValue("item"); item != "" {
		err = redisClient.LRem(removeKey, 1, item).Err()
	}

	return
}

func retryAllHandler(req *http.Request, rep *httpReply) (err error) {
	path := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(path) < 3 {
		err = fmt.Errorf("Invalid path")
		return
	}

	listType := path[1]
	queueName := path[2]

	removeKey := fmt.Sprintf("%s:queue:%s:%s", redisHeader, queueName, listType)
	pushKey := fmt.Sprintf("%s:queue:%s:pending", redisHeader, queueName)

	err = myredis.FlushList(removeKey, pushKey)
	return
}

func deleteAllHandler(req *http.Request, rep *httpReply) (err error) {
	path := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(path) < 3 {
		err = fmt.Errorf("Invalid path")
		return
	}

	listType := path[1]
	queueName := path[2]

	removeKey := fmt.Sprintf("%s:queue:%s:%s", redisHeader, queueName, listType)
	err = redisClient.Del(removeKey).Err()
	return
}
