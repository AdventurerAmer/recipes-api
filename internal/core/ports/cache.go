package ports

import (
	"context"
	"time"

	"github.com/AdventurerAmer/recipes-api/errs"
)

type Cache interface {
	Get(ctx context.Context, key string, v any) error
	Put(ctx context.Context, key string, v any, TTL time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type cacheStub struct{}

func (c *cacheStub) Get(ctx context.Context, key string, v any) error {
	return errs.NewNotFound(nil, "key not found")
}

func (c *cacheStub) Put(ctx context.Context, key string, v any, TTL time.Duration) error {
	return nil
}

func (c *cacheStub) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func NewCacheStub() Cache {
	return &cacheStub{}
}
