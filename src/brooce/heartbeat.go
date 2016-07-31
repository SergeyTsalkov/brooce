package main

import (
	"log"
	"time"
)

func heartbeater() {
	for {
		heartbeat()
		time.Sleep(time.Minute)
	}
}

func heartbeat() {
	key := heartbeatKey + ":" + myProcName
	err := redisClient.Set(key, "1", 15*time.Minute).Err()
	if err != nil {
		log.Println("redis heartbeat error:", err)
	}

}
