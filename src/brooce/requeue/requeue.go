package requeue

import (
	"log"
	"strings"

	"brooce/config"
	myredis "brooce/redis"
	"brooce/util"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

func Start() {
	go func() {
		for {
			util.SleepUntilNextMinute()
			err := requeue()
			if err != nil {
				log.Println("Error trying to requeue delayed jobs:", err)
			}
		}
	}()
}

func requeue() (err error) {
	var keys []string
	keys, err = redisClient.Keys(redisHeader + ":queue:*:delayed").Result()
	if err != nil {
		return
	}

	for _, delayedKey := range keys {
		pendingKey := strings.TrimSuffix(delayedKey, ":delayed") + ":pending"
		err = myredis.FlushList(delayedKey, pendingKey)
		if err != nil {
			return
		}
	}

	return
}
