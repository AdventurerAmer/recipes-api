package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/errs"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) ports.Cache {
	return &redisCache{
		client: client,
	}
}

func (rc *redisCache) Get(ctx context.Context, key string, v any) error {
	data, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return errs.NewNotFound(nil, "key not found")
	} else if err != nil {
		return fmt.Errorf("'client.Get' failed: %w", err)
	}

	if err := json.Unmarshal([]byte(data), v); err != nil {
		return fmt.Errorf("'json.Unmarshal' failed: %w", err)
	}
	return nil
}

func (rc *redisCache) GetVersioned(ctx context.Context, key, versionKey string, v any) error {
	script := `
        local version = redis.call('GET', KEYS[1])
		
		if not version then
			version = 0
		end
		
		local versionedKey = KEYS[2] .. ':version:' .. version
		return redis.call('GET', versionedKey)
    `
	data, err := rc.client.Eval(ctx, script, []string{versionKey, key}).Result()
	if err == redis.Nil {
		return errs.NewNotFound(nil, "key not found")
	} else if err != nil {
		return fmt.Errorf("'client.Get' failed: %w", err)
	}
	if err := json.Unmarshal([]byte(data.(string)), v); err != nil {
		return fmt.Errorf("'json.Unmarshal' failed: %w", err)
	}
	return nil
}

func (rc *redisCache) Put(ctx context.Context, key string, v any, TTL time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}
	if _, err := rc.client.Set(ctx, key, data, TTL).Result(); err != nil {
		return fmt.Errorf("'client.Set' failed: %w", err)
	}
	return nil
}

func (rc *redisCache) PutVersioned(ctx context.Context, key string, versionKey string, v any, TTL time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}

	script := `
        local version = redis.call('GET', KEYS[1])
		
		if not version then
			version = 0
		end
		
		local versionedKey = KEYS[2] .. ':version:' .. version
		return redis.call('set', versionedKey, ARGV[1], 'EX', tonumber(ARGV[2]))
    `
	if _, err := rc.client.Eval(ctx, script, []string{versionKey, key}, string(data), int(TTL.Seconds())).Result(); err != nil {
		return fmt.Errorf("'client.Set' failed: %w", err)
	}
	return nil
}

func (rc *redisCache) Delete(ctx context.Context, keys ...string) error {
	if _, err := rc.client.Del(ctx, keys...).Result(); err != nil {
		return fmt.Errorf("'client.Del' failed: %w", err)
	}
	return nil
}

func (rc *redisCache) Inc(ctx context.Context, key string) (int64, error) {
	value, err := rc.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("'client.Incr' failed: %w", err)
	}
	return value, nil
}
