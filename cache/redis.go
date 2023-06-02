package cache

import (
	"context"
	"time"
)

type PostCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	Delete(ctx context.Context, key string) error
}

type Cache struct {
	PostCache
}

func NewCache(host string, password string, db int, exp time.Duration) *Cache {
	return &Cache{
		PostCache: NewRedisCache(host, password, db, exp),
	}
}
