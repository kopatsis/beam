package config

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

func RateLimit(client *redis.Client, key string, maxRequests int, keyType string) error {
	ctx := context.Background()

	windowKey := fmt.Sprintf("WDNW::%s::%s", keyType, key)

	count, err := client.LLen(ctx, windowKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get window length: %w", err)
	}

	if count >= int64(maxRequests) {
		return fmt.Errorf("rate limit exceeded for key: %s", key)
	}

	now := time.Now().Unix()
	_, err = client.LPush(ctx, windowKey, now).Result()
	if err != nil {
		return fmt.Errorf("failed to push timestamp: %w", err)
	}

	_, err = client.LTrim(ctx, windowKey, 0, int64(maxRequests-1)).Result()
	if err != nil {
		return fmt.Errorf("failed to trim old timestamps: %w", err)
	}

	err = client.Expire(ctx, windowKey, time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}
