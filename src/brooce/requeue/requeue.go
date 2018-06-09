package requeue

import (
	"fmt"
	"log"

	"brooce/config"
	"brooce/listing"
	myredis "brooce/redis"
	"brooce/util"
)

var redisHeader = config.Config.ClusterName
var requeueInterval = config.Config.Requeue.Interval

func Start() {
	go func() {
		for {
			util.SleepUntilNextInterval(requeueInterval)
			err := requeue()
			if err != nil {
				log.Println("Error trying to requeue delayed jobs:", err)
			}
		}
	}()
}

func requeue() (err error) {
	var queues map[string]*listing.QueueInfoType
	queues, err = listing.Queues(true)
	if err != nil {
		return
	}

	for name, _ := range queues {
		pendingKey := fmt.Sprintf("%s:queue:%s:pending", redisHeader, name)
		delayedKey := fmt.Sprintf("%s:queue:%s:delayed", redisHeader, name)

		err = myredis.FlushList(delayedKey, pendingKey)
		if err != nil {
			return
		}
	}

	return
}
