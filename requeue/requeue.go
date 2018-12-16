package requeue

import (
	"log"

	"brooce/config"
	"brooce/listing"
	myredis "brooce/redis"
	"brooce/util"
)

var redisHeader = config.Config.ClusterName

func Start() {
	queues, err := listing.Queues(true)
	if err != nil {
		log.Fatalln("Unable to list running queues:", err)
	}

	for _, queue := range queues {
		opts := queue.JobOptions()

		go requeue(queue, queue.DelayedList(), opts.RequeueDelayed())
		go requeue(queue, queue.FailedList(), opts.RequeueFailed())
	}
}

func requeue(queue *listing.QueueInfoType, listToRequeue string, interval int) {
	if (listToRequeue == queue.FailedList() && interval > 0) || (listToRequeue == queue.DelayedList() && interval != 60) {
		log.Println("Will requeue", listToRequeue, "to", queue.PendingList(), "every", interval, "seconds")
	}

	if interval == 0 {
		return
	}

	for {
		util.SleepUntilNextInterval(interval)

		log.Println("Requeued", listToRequeue, "to", queue.PendingList())
		err := myredis.FlushList(listToRequeue, queue.PendingList())
		if err != nil {
			log.Println("Failed to requeue", listToRequeue, "to", queue.PendingList(), ":", err)
		}
	}

}
