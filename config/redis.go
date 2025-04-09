package config

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379" // fallback for local dev
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}
