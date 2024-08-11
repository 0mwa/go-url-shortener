package database

import (
	"github.com/go-redis/redis/v8"
	"os"
)

func NewRedisClient(dbNo int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       dbNo,
	})
	return rdb
}
