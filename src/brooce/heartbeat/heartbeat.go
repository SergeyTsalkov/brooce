package heartbeat

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"brooce/config"
	"brooce/myip"
	myredis "brooce/redis"
)

var heartbeatEvery = 30 * time.Second
var assumeDeadAfter = 95 * time.Second

var redisClient = myredis.Get()
var heartbeatStr = makeHeartbeat()
var once sync.Once

type HeartbeatType struct {
	ProcName string         `json:"procname"`
	Hostname string         `json:"hostname"`
	IP       string         `json:"ip"`
	PID      int            `json:"pid"`
	Queues   map[string]int `json:"queues"`
}

func (hb *HeartbeatType) TotalThreads() (total int) {
	if hb.Queues == nil {
		return
	}

	for _, ct := range hb.Queues {
		total += ct
	}
	return
}

func makeHeartbeat() string {
	hb := &HeartbeatType{
		ProcName: config.Config.ProcName,
		IP:       myip.PublicIPv4(),
		PID:      os.Getpid(),
		Queues:   config.Config.Queues,
	}

	var err error
	hb.Hostname, err = os.Hostname()
	if err != nil {
		log.Println("Warning: Unable to determine machine hostname!")
	}

	bytes, err := json.Marshal(hb)
	if err != nil {
		log.Fatalln(err)
	}

	return string(bytes)
}

func Start() {
	// need to send a single heartbeat FOR SURE before we grab a job!
	heartbeat()

	go func() {
		for {
			time.Sleep(heartbeatEvery)
			heartbeat()
		}
	}()
}

func heartbeat() {
	/*once.Do(func() {
		value = makeHeartbeat()
	})*/

	key := fmt.Sprintf("%s:workerprocs:%s", config.Config.ClusterName, config.Config.ProcName)
	err := redisClient.Set(key, heartbeatStr, assumeDeadAfter).Err()
	if err != nil {
		log.Println("redis heartbeat error:", err)
	}

}
