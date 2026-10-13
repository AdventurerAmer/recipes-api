package ports

import (
	"context"
	"time"

	"github.com/AdventurerAmer/recipes-api/errs"
)

type Cache interface {
	Get(ctx context.Context, key string, v any) error
	GetVersioned(ctx context.Context, key, versionKey string, v any) (int64, bool, error)
	Put(ctx context.Context, key string, v any, TTL time.Duration) error
	PutVersioned(ctx context.Context, key, versionKey string, v any, TTL time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	IncDelPrev(ctx context.Context, keys ...string) error
}

type cacheStub struct{}

func (c *cacheStub) Get(ctx context.Context, key string, v any) error {
	return errs.NewNotFound(nil, "key not found")
}

func (c *cacheStub) GetVersioned(ctx context.Context, key, versionKey string, v any) (int64, bool, error) {
	return 0, false, nil
}

func (c *cacheStub) Put(ctx context.Context, key string, v any, TTL time.Duration) error {
	return nil
}

func (c *cacheStub) PutVersioned(ctx context.Context, key, versionKey string, v any, TTL time.Duration) error {
	return nil
}

func (c *cacheStub) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (c *cacheStub) IncDelPrev(ctx context.Context, keys ...string) error {
	return nil
}

func NewCacheStub() Cache {
	return &cacheStub{}
}
