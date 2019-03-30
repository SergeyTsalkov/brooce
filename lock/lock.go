package lock

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"brooce/config"
	"brooce/heartbeat"
	"brooce/listing"
	myredis "brooce/redis"

	"github.com/go-redis/redis"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

func GrabLocks(locks []string, actor string) (success bool, err error) {
	if len(locks) == 0 {
		return true, nil
	}

	results := make([]*redis.IntCmd, len(locks))

	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for i, lock := range locks {
			results[i] = pipe.LPush(lockRedisKey(lock), actor)
		}
		return nil
	})
	if err != nil {
		return
	}

	for i, result := range results {
		if result.Val() > lockDepth(locks[i]) {
			err = ReleaseLocks(locks, actor)
			return
		}
	}

	success = true
	return
}

func ReleaseLocks(locks []string, actor string) (err error) {
	if len(locks) == 0 {
		return
	}

	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for _, lock := range locks {
			pipe.LRem(lockRedisKey(lock), 0, actor)
		}
		return nil
	})

	return
}

func Start() {
	// before we grab a job, cleanup any lingering locks for our own threads
	err := cleanupOwn()
	if err != nil {
		log.Fatalln("redis error:", err)
	}

	go func() {
		for {
			err = cleanupAll()
			if err != nil {
				log.Println("redis error:", err)
			}

			time.Sleep(time.Minute)
		}
	}()

}

func cleanupOwn() (err error) {
	var keys []string
	keys, err = myredis.ScanKeys(redisHeader + ":lock:*")
	if err != nil || len(keys) == 0 {
		return
	}

	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for _, thread := range config.Threads {
			for _, key := range keys {
				log.Println("Pruning orphaned lock", thread.Name, "from", key)
				pipe.LRem(key, 0, thread.Name)
			}
		}
		return nil
	})

	return
}

func cleanupAll() (err error) {
	var lockKeys []string
	lockKeys, err = myredis.ScanKeys(redisHeader + ":lock:*")
	if err != nil || len(lockKeys) == 0 {
		return
	}

	lrangeResults := make([]*redis.StringSliceCmd, len(lockKeys))
	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for i, key := range lockKeys {
			lrangeResults[i] = pipe.LRange(key, 0, -1)
		}
		return nil
	})
	if err != nil || len(lrangeResults) == 0 {
		return
	}

	actors := map[string]bool{}
	for _, result := range lrangeResults {
		for _, actor := range result.Val() {
			actors[actor] = true
		}
	}

	var workers []*heartbeat.HeartbeatType
	workers, err = listing.RunningWorkers()
	if err != nil {
		return
	}

	for _, worker := range workers {
		for _, thread := range worker.Threads {
			delete(actors, thread.Name)
		}
	}

	if len(actors) == 0 {
		return
	}

	_, err = redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for actor := range actors {
			for _, key := range lockKeys {
				log.Println("Pruning orphaned lock", actor, "from", key)
				pipe.LRem(key, 0, actor)
			}
		}
		return nil
	})

	return
}

func lockRedisKey(lock string) string {
	return fmt.Sprintf("%s:lock:%s", redisHeader, lock)
}

func lockDepth(lock string) int64 {
	if parts := strings.Split(lock, ":"); len(parts) >= 2 {
		if depth, err := strconv.Atoi(parts[0]); err == nil {
			return int64(depth)
		}
	}

	return 1
}
