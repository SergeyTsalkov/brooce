// +build cluster

package redis

import (
	"log"
	"strings"
	"sync"
	"time"

	"brooce/config"

	"github.com/go-redis/redis"
)

var redisClient *redis.ClusterClient
var once sync.Once

func Get() *redis.ClusterClient {
	once.Do(func() {
		threads := len(config.Threads) + 10

		redisClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        strings.Split(config.Config.Redis.Host, ","),
			Password:     config.Config.Redis.Password,
			MaxRetries:   10,
			PoolSize:     threads,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 5 * time.Second,
			PoolTimeout:  1 * time.Second,
		})

		for {
			err := redisClient.Ping().Err()
			if err == nil {
				break
			}
			log.Println("Can't reach redis at", config.Config.Redis.Host, "-- are your redis addr and password right?", err)
			time.Sleep(5 * time.Second)
		}
	})

	return redisClient
}

func FlushList(src, dst string) (err error) {
	redisClient := Get()
	for err == nil {
		_, err = redisClient.RPopLPush(src, dst).Result()
	}

	if err == redis.Nil {
		err = nil
	}

	return
}

func ScanKeys(match string) (keys []string, err error) {
	redisClient := Get()
	cursor := uint64(0)

	for {
		var result []string
		result, cursor, err = redisClient.Scan(cursor, match, 10000).Result()

		if err != nil {
			return
		}

		keys = append(keys, result...)

		if cursor == 0 {
			break
		}
	}

	return
}
