package ports

import (
	"context"
	"time"

	"github.com/AdventurerAmer/recipes-api/errs"
)

type Cache interface {
	Get(ctx context.Context, key string, v any) error
	GetVersioned(ctx context.Context, key, versionKey string, v any) error
	Put(ctx context.Context, key string, v any, TTL time.Duration) error
	PutVersioned(ctx context.Context, key string, versionKey string, v any, TTL time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Inc(ctx context.Context, key string) (int64, error)
}

type cacheStub struct{}

func (c *cacheStub) Get(ctx context.Context, key string, v any) error {
	return errs.NewNotFound(nil, "key not found")
}

func (c *cacheStub) GetVersioned(ctx context.Context, key, versionKey string, v any) error {
	return errs.NewNotFound(nil, "key not found")
}

func (c *cacheStub) Put(ctx context.Context, key string, v any, TTL time.Duration) error {
	return nil
}

func (c *cacheStub) PutVersioned(ctx context.Context, key string, versionKey string, v any, TTL time.Duration) error {
	return nil
}

func (c *cacheStub) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (c *cacheStub) Inc(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func NewCacheStub() Cache {
	return &cacheStub{}
}
