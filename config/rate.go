package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// Under max requests, error
func RateLimit(client *redis.Client, store, classification, trueKey string, maxRequests int, period time.Duration) (bool, error) {
	ctx := context.Background()

	windowKey := fmt.Sprintf("%s::WNDW::%s", store, classification)
	if trueKey != "" {
		windowKey += "::" + trueKey
	}

	now := time.Now()
	startTime := now.Add(-period).Unix()

	_, err := client.ZRemRangeByScore(ctx, windowKey, "0", fmt.Sprintf("%d", startTime)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to remove old timestamps: %w", err)
	}

	count, err := client.ZCard(ctx, windowKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to get the number of requests: %w", err)
	}

	if count >= int64(maxRequests) {
		return false, nil
	}

	_, err = client.ZAdd(ctx, windowKey, &redis.Z{
		Score: float64(now.Unix()),
	}).Result()
	if err != nil {
		return false, fmt.Errorf("failed to add timestamp: %w", err)
	}

	err = client.Expire(ctx, windowKey, period).Err()
	if err != nil {
		return false, fmt.Errorf("failed to set expiration: %w", err)
	}

	return true, nil
}

func IsValidDeviceID(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func BotDetection(c *gin.Context) error {
	ua := c.GetHeader("User-Agent")

	if ua == "" {
		return fmt.Errorf("Blocked")
	}

	blockedKeywords := []string{
		"headless", "phantomjs", "selenium", "puppeteer", "scrapy", "curl", "wget", "python-requests",
		"java", "httpclient", "urllib", "node-fetch", "go-http-client",
	}

	loweredUA := strings.ToLower(ua)
	for _, keyword := range blockedKeywords {
		if strings.Contains(loweredUA, keyword) {
			return fmt.Errorf("blocked (%s)", ua)
		}
	}

	return nil
}
