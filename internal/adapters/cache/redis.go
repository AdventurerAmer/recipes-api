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

func (rc *redisCache) GetVersioned(ctx context.Context, key, versionKey string, v any) (int64, bool, error) {
	script := `
        local version = redis.call('GET', KEYS[1])
		
		if not version then
			version = "0"
		end
		
		local versionedKey = KEYS[2] .. ':version:' .. version
		return { tonumber(version), redis.call('GET', versionedKey) }
    `
	results, err := rc.client.Eval(ctx, script, []string{versionKey, key}).Slice()
	if err != nil {
		return 0, false, fmt.Errorf("'client.Get' failed: %w", err)
	}
	version := results[0].(int64)
	if results[1] == nil {
		return version, false, nil
	}
	data := results[1].(string)
	if err := json.Unmarshal([]byte(data), v); err != nil {
		return 0, false, fmt.Errorf("'json.Unmarshal' failed: %w", err)
	}
	return version, true, nil
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

func (rc *redisCache) PutVersioned(ctx context.Context, key, versionKey string, v any, TTL time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}

	script := `
        local version = redis.call('GET', KEYS[2])
		
		if not version then
			version = "0"
		end

		local lastPart = KEYS[1]:match("[^:]+$")
		local expectedVersion = lastPart
		if version ~= expectedVersion then
			return redis.status_reply("OK")
		end
		
		return redis.call('SET', KEYS[1], ARGV[1], 'EX', tonumber(ARGV[2]))
    `
	if _, err := rc.client.Eval(ctx, script, []string{key, versionKey}, string(data), int(TTL.Seconds())).Result(); err != nil {
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

func (rc *redisCache) IncDelPrev(ctx context.Context, keys ...string) error {
	script := `
		for i = 1, #KEYS do
			local val = tonumber(redis.call("INCR", KEYS[i]))
			local old = val - 1
			local key = KEYS[i] .. ":" .. old
			redis.call("DEL", key)
		end
		return redis.status_reply("OK")
	`
	if _, err := rc.client.Eval(ctx, script, keys).Result(); err != nil {
		return fmt.Errorf("'client.Eval' failed: %w", err)
	}
	return nil
}
