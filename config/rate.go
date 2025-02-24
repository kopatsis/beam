package config

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

func RateLimit(client *redis.Client, store, classification, trueKey string, maxRequests int, period time.Duration) error {
	ctx := context.Background()

	windowKey := fmt.Sprintf("%s::WNDW::%s", store, classification)
	if trueKey != "" {
		windowKey += "::" + trueKey
	}

	now := time.Now()
	startTime := now.Add(-period).Unix()

	_, err := client.ZRemRangeByScore(ctx, windowKey, "0", fmt.Sprintf("%d", startTime)).Result()
	if err != nil {
		return fmt.Errorf("failed to remove old timestamps: %w", err)
	}

	count, err := client.ZCard(ctx, windowKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get the number of requests: %w", err)
	}

	if count >= int64(maxRequests) {
		return fmt.Errorf("rate limit exceeded for key: %s", windowKey)
	}

	_, err = client.ZAdd(ctx, windowKey, &redis.Z{
		Score: float64(now.Unix()),
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to add timestamp: %w", err)
	}

	err = client.Expire(ctx, windowKey, time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}
