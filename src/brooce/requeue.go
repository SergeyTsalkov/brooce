package main

import "strings"

var autoRequeueHash = map[string]bool{}

func autoRequeueDelayed(queue string) {
	// only run on each queue once
	if _, ok := autoRequeueHash[queue]; ok {
		return
	}
	autoRequeueHash[queue] = true

	for {
		sleepUntil00()
		requeueDelayed(queue)
	}
}

func requeueDelayed(queue string) {
	pendingList := strings.Join([]string{redisHeader, "queue", queue, "pending"}, ":")
	delayedList := strings.Join([]string{redisHeader, "queue", queue, "delayed"}, ":")

	var err error
	for err == nil {
		_, err = redisClient.RPopLPush(delayedList, pendingList).Result()
	}
}
