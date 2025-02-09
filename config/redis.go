package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient() *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	fmt.Println("Connected to Redis successfully")
	return rdb
}

type FlushKey struct {
	ActualKey string    `json:"a"`
	CanFlush  time.Time `json:"c"`
}

func (f *FlushKey) Set() *FlushKey {
	f.ActualKey = strconv.FormatInt(time.Now().Unix(), 10)
	f.CanFlush = time.Now().Add(time.Duration(BATCH) * time.Second)
	return f
}

func (f *FlushKey) ToJSON() string {
	data, _ := json.Marshal(f)
	return string(data)
}

func NewKey() string {
	return (&FlushKey{}).Set().ToJSON()
}
