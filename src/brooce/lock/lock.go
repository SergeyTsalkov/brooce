package lock

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"brooce/config"
	myredis "brooce/redis"

	redis "gopkg.in/redis.v5"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

func GrabLocks(locks []string) (success bool, err error) {
	actor := config.Config.ProcName
	if len(locks) == 0 {
		return true, nil
	}

	results := make([]*redis.IntCmd, len(locks))

	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
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
			err = ReleaseLocks(locks)
			return
		}
	}

	success = true
	return
}

func ReleaseLocks(locks []string) (err error) {
	actor := config.Config.ProcName

	if len(locks) == 0 {
		return
	}

	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for _, lock := range locks {
			pipe.LRem(lockRedisKey(lock), 1, actor)
		}
		return nil
	})

	return
}

func Start() {
	err := cleanup(config.Config.ProcName)
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

func cleanup(actor string) (err error) {
	var keys []string
	keys, err = redisClient.Keys(redisHeader + ":lock:*").Result()
	if err != nil || len(keys) == 0 {
		return
	}

	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for _, key := range keys {
			pipe.LRem(key, 0, actor)
		}
		return nil
	})

	return
}

func cleanupAll() (err error) {
	var lockKeys []string
	lockKeys, err = redisClient.Keys(redisHeader + ":lock:*").Result()
	if err != nil || len(lockKeys) == 0 {
		return
	}

	lrangeResults := make([]*redis.StringSliceCmd, len(lockKeys))
	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
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

	var workerKeys []string
	workerKeys, err = redisClient.Keys(redisHeader + ":workerprocs:*").Result()
	if err != nil {
		return
	}
	for _, key := range workerKeys {
		parts := strings.SplitN(key, ":", 3)
		if len(parts) < 3 {
			continue
		}

		delete(actors, parts[2])
	}

	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for actor, _ := range actors {
			for _, key := range lockKeys {
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
