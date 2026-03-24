package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type AlertCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewAlertCache(client *redis.Client) *AlertCache {
	return &AlertCache{
		client: client,
		ttl:    30 * time.Minute,
	}
}

func (c *AlertCache) GetFingerprint(ctx context.Context, fingerprint string) (string, error) {
	key := fmt.Sprintf("alert:fp:%s", fingerprint)
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("not found")
	}
	return val, err
}

func (c *AlertCache) SetFingerprint(ctx context.Context, fingerprint string, alertID string) error {
	key := fmt.Sprintf("alert:fp:%s", fingerprint)
	return c.client.Set(ctx, key, alertID, c.ttl).Err()
}

func (c *AlertCache) IncrFlapping(ctx context.Context, fingerprint string) (int64, error) {
	key := fmt.Sprintf("alert:flap:%s", fingerprint)
	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 30*time.Minute)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}
