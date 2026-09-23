package infra

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	*config.Redis
}

func (cfg *Redis) Connect(ctx context.Context) (Disconnecter, error) {
	opts := &redis.Options{
		Addr:     cfg.Addr(),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       *cfg.Database,
	}
	client := redis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("'client.Ping' failed: %w", err)
	}
	return &RedisContext{Client: client}, nil
}

type RedisContext struct {
	Client *redis.Client
}

func (c *RedisContext) Disconnect(ctx context.Context) error {
	errCh := make(chan error)
	go func() {
		if err := c.Client.Close(); err != nil {
			errCh <- fmt.Errorf("'Client.Close' failed: %w", err)
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
