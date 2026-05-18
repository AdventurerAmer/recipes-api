package infra

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Address  string `cfg:"address"`
	Username string `cfg:"username"`
	Password string `cfg:"password"`
	Database int    `cfg:"database"`
}

func (cfg *RedisConfig) Connect(ctx context.Context) (Disconnecter, error) {
	opts := &redis.Options{
		Addr:     cfg.Address,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.Database,
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
