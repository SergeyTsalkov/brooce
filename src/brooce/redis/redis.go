package redis

import (
	"sync"
	"time"

	"brooce/config"

	redis "gopkg.in/redis.v3"
)

var redisClient *redis.Client
var once sync.Once

func Get() *redis.Client {
	once.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:         config.Config.Redis.Host,
			Password:     config.Config.Redis.Password,
			MaxRetries:   2,
			PoolSize:     10, // TODO: make this equal to number of threads
			DialTimeout:  5 * time.Second,
			ReadTimeout:  20 * time.Second,
			WriteTimeout: 5 * time.Second,
			PoolTimeout:  1 * time.Second,
		})
	})

	return redisClient
}
