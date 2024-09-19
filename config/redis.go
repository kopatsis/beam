package config

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient() *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	if redisAddr == "" || redisPassword == "" {
		log.Fatalf("missing a variable in .env to set up Redis")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	return client
}
