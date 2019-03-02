package requeue

import (
	"log"

	"brooce/config"
	myredis "brooce/redis"
	"brooce/util"
)

var redisHeader = config.Config.ClusterName

func Start() {
	for _, queue := range config.Config.Queues {
		opts := queue.DeepJobOptions()

		go requeue(queue, queue.DelayedList(), opts.RequeueDelayed())
		go requeue(queue, queue.FailedList(), opts.RequeueFailed())
	}
}

func requeue(queue config.Queue, listToRequeue string, interval int) {
	if (listToRequeue == queue.FailedList() && interval > 0) || (listToRequeue == queue.DelayedList() && interval != 60) {
		log.Println("Will requeue", listToRequeue, "to", queue.PendingList(), "every", interval, "seconds")
	}

	if interval == 0 {
		return
	}

	for {
		util.SleepUntilNextInterval(interval)

		//log.Println("Requeued", listToRequeue, "to", queue.PendingList())
		err := myredis.FlushList(listToRequeue, queue.PendingList())
		if err != nil {
			log.Println("Failed to requeue", listToRequeue, "to", queue.PendingList(), ":", err)
		}
	}

}
