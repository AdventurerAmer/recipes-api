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

type RedisContext struct {
	Client *redis.Client
}

func ConnectToRedis(ctx context.Context, cfg RedisConfig) (RedisContext, error) {
	opts := &redis.Options{
		Addr:     cfg.Address,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.Database,
	}
	client := redis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		return RedisContext{}, fmt.Errorf("'client.Ping' failed: %w", err)
	}
	return RedisContext{Client: client}, nil
}

func disconnectFromRedis(ctx context.Context, redisCtx RedisContext) error {
	errCh := make(chan error)
	go func() {
		if err := redisCtx.Client.Close(); err != nil {
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
