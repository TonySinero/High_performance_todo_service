package cache

import (
	"context"

	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client  *redis.Client
	expires time.Duration
}

func NewRedisCache(host, password string, db int, exp time.Duration) *RedisCache {
	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     host,
			Password: password,
			DB:       db,
		}),
		expires: exp,
	}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	cmd := c.client.Get(ctx, key)

	// Check if key exists
	if cmd.Err() == redis.Nil {
		return "", fmt.Errorf("redis: key '%s' not found", key)
	} else if cmd.Err() != nil {
		return "", fmt.Errorf("redis: error occurred while getting key '%s' - %v", key, cmd.Err())
	}

	return cmd.Val(), nil
}

func (c *RedisCache) Set(ctx context.Context, key, value string) error {
	err := c.client.Set(ctx, key, value, c.expires).Err()
	if err != nil {
		return fmt.Errorf("redis: error occurred while setting key '%s' - %v", key, err)
	}
	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis: error occurred while deleting key '%s' - %v", key, err)
	}
	return nil
}
