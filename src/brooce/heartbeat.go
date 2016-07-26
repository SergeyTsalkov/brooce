package main

import "time"

func heartbeater() {
	for {
		err := heartbeat()
		if err != nil {
			logger.Println("heartbeat error:", err)
		}

		time.Sleep(time.Minute)
	}
}

func heartbeat() error {
	key := heartbeatKey + ":" + myProcName
	return redisClient.Set(key, "1", 15*time.Minute).Err()
}
