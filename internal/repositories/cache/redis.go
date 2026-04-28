package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisWrapper struct {
	client *redis.Client
	ttl    time.Duration
}

func (r redisWrapper) get(ctx context.Context, key string, value any) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("'client.Get' failed: %w", err)
	}
	if err := json.Unmarshal([]byte(data), value); err != nil {
		return fmt.Errorf("'json.Unmarshal' failed: %w", err)
	}
	return nil
}

func (r redisWrapper) set(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}
	if _, err := r.client.Set(ctx, key, data, r.ttl).Result(); err != nil {
		return fmt.Errorf("'client.Set' failed: %w", err)
	}
	return nil
}

func (r redisWrapper) del(ctx context.Context, keys ...string) error {
	if _, err := r.client.Del(ctx, keys...).Result(); err != nil {
		return fmt.Errorf("'client.Del' failed: %w", err)
	}
	return nil
}
