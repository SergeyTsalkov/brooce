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
	"brooce/util"
)

var heartbeatEvery = 30 * time.Second
var assumeDeadAfter = 95 * time.Second

var redisClient = myredis.Get()
var once sync.Once

type HeartbeatType struct {
	ProcName  string              `json:"procname"`
	Hostname  string              `json:"hostname"`
	IP        string              `json:"ip"`
	PID       int                 `json:"pid"`
	Timestamp int64               `json:"timestamp"`
	Threads   []config.ThreadType `json:"threads"`
}

func (hb *HeartbeatType) HeartbeatAge() time.Duration {
	return time.Since(time.Unix(hb.Timestamp, 0))
}

func (hb *HeartbeatType) HeartbeatTooOld() bool {
	return hb.HeartbeatAge() > assumeDeadAfter
}

// if heartbeat is for worker on the same machine, we can determine
// if the PID corresponds to a running process
func (hb *HeartbeatType) IsLocalZombie() bool {
	if hb.IP != myip.PublicIPv4() {
		return false
	}

	if hb.PID == 0 || hb.PID == os.Getpid() {
		return false
	}

	return !util.ProcessExists(hb.PID)
}

func (hb *HeartbeatType) Queues() (queues map[string]int) {
	queues = map[string]int{}

	for _, thread := range hb.Threads {
		queues[thread.Queue] += 1
	}

	return
}

func makeHeartbeat() string {
	hb := &HeartbeatType{
		ProcName:  config.Config.ProcName,
		IP:        myip.PublicIPv4(),
		PID:       os.Getpid(),
		Timestamp: time.Now().Unix(),
		Threads:   config.Threads,
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
	key := fmt.Sprintf("%s:workerprocs", config.Config.ClusterName)
	err := redisClient.HSet(key, config.Config.ProcName, makeHeartbeat()).Err()
	if err != nil {
		log.Println("redis heartbeat error:", err)
	}
}
