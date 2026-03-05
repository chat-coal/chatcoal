package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var sfGroup singleflight.Group

var Redis *redis.Client

// logRedisErr logs a warning when a Redis command returns an error.
// Use as: logRedisErr(Redis.Set(...).Err(), "tag")
func logRedisErr(err error, tag string) {
	if err != nil {
		log.Warnf("[cache] %s: %v", tag, err)
	}
}

// AnonMessageRateLimitOK returns true if the anonymous user is within the 2 msg/sec limit.
func AnonMessageRateLimitOK(userID int64) bool {
	if Redis == nil {
		return true
	}
	key := fmt.Sprintf("anon:rl:%d", userID)
	count, err := Redis.Incr(context.Background(), key).Result()
	if err != nil {
		return true
	}
	if count == 1 {
		Redis.Expire(context.Background(), key, time.Second)
	}
	return count <= 2
}

func Connect() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	Redis = redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     50,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	if err := Redis.Ping(context.Background()).Err(); err != nil {
		log.Warn("Redis unavailable, caching disabled: ", err)
		Redis = nil
		return nil
	}

	log.Info("Redis connected at ", addr)
	return nil
}
