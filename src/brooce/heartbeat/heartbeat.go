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

	redis "gopkg.in/redis.v3"
)

var HeartbeatEvery = 30 * time.Second
var AssumeDeadAfter = 95 * time.Second
var removeDeadAfter = 7 * 24 * time.Hour

var redisClient = myredis.Get()
var once sync.Once

type HeartbeatType struct {
	ProcName string         `json:"procname"`
	Hostname string         `json:"hostname"`
	IP       string         `json:"ip"`
	PID      int            `json:"pid"`
	Queues   map[string]int `json:"queues"`
	TS       int64          `json:"timestamp"`
}

type HeartbeatTemplateType struct {
	*HeartbeatType
	StatusColor    string   `json:"status_color"`
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
		TS:       time.Now().Unix(),
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
	auditHeartbeats()

	go func() {
		for {
			time.Sleep(HeartbeatEvery)
			heartbeat()
			auditHeartbeats()
		}
	}()
}

func heartbeat() {
	heartbeatStr := makeHeartbeat()
	key := fmt.Sprintf("%s:workerprocs:%s", config.Config.ClusterName, config.Config.ProcName)
	err := redisClient.Set(key, heartbeatStr, removeDeadAfter).Err()
	if err != nil {
		log.Println("redis heartbeat error:", err)
	}
}

// check other processes on same IP, make sure they're actually there
func auditHeartbeats() {
	var err error
	defer func() {
		if err != nil {
			log.Println("redis audit heartbeat error:", err)
		}
	}()

	keyMatch := fmt.Sprintf("%s:workerprocs:%v-*", config.Config.ClusterName, myip.PublicIPv4())
	var keys []string
	keys, err = redisClient.Keys(keyMatch).Result()
	if err != nil {
		return
	}

	heartbeats := map[string]*redis.StringCmd{}
	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for _, key := range keys {
			result := pipe.Get(key)
			heartbeats[key] = result
		}
		return nil
	})
	if err != nil {
		return
	}

	for key, str := range heartbeats {
		worker := &HeartbeatTemplateType{}
		err = json.Unmarshal([]byte(str.Val()), worker)
		if err != nil {
			return
		}

		if worker.PID == 0 || worker.PID == os.Getpid() {
			continue
		}

		if !util.ProcessExists(worker.PID) {
			log.Printf("Purging dead worker, was PID %v", worker.PID)
			err = redisClient.Del(key).Err()
			if err != nil {
				return
			}
		} else if IsAlive(worker) == 1 {
			log.Println("Warning: Running multiple instances of brooce on the same machine is not recommended. Use threads in one instance instead!")
		}

	}
}

func IsAlive(worker *HeartbeatTemplateType) int {
	workerTS := time.Unix(worker.TS, 0)
	currentTS := time.Now().Unix()

	if currentTS > workerTS.Add(AssumeDeadAfter).Unix() {
		worker.StatusColor = "red"
		return -1
	} else if currentTS < workerTS.Add(AssumeDeadAfter).Unix() && currentTS > workerTS.Add(HeartbeatEvery).Unix() {
		worker.StatusColor = "yellow"
		return 0
	} else if currentTS <= workerTS.Add(HeartbeatEvery).Unix() {
		worker.StatusColor = "green"
		return 1
	} else {
		worker.StatusColor = "grey"
		return -11
	}
}