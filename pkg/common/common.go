package common

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	RedisAddr string
}

func LoadConfig() *Config {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return &Config{
		RedisAddr: addr,
	}
}

func NewRedisClient(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return rdb, nil
}
