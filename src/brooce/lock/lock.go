package lock

import (
	"fmt"
	"strconv"
	"strings"

	"brooce/config"
	myredis "brooce/redis"

	redis "gopkg.in/redis.v3"
)

var redisClient = myredis.Get()
var redisHeader = config.Config.ClusterName

func GrabLocks(actor string, locks []string) (success bool, err error) {
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
			err = ReleaseLocks(actor, locks)
			return
		}
	}

	success = true
	return
}

func ReleaseLocks(actor string, locks []string) (err error) {
	_, err = redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		for _, lock := range locks {
			pipe.LRem(lockRedisKey(lock), 1, actor)
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
