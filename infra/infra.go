package infra

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Infra struct {
	StartupTimeout  time.Duration
	ShutdownTimeout time.Duration
	Mongo           map[MongoConfig]*MongoContext
	Redis           map[RedisConfig]*RedisContext
}

func New() *Infra {
	return &Infra{
		StartupTimeout:  time.Second,
		ShutdownTimeout: time.Second,
		Mongo:           make(map[MongoConfig]*MongoContext),
		Redis:           make(map[RedisConfig]*RedisContext),
	}
}

func (infra *Infra) BindMongo(cfg MongoConfig, ctx *MongoContext) {
	infra.Mongo[cfg] = ctx
}

func (infra *Infra) BindRedis(cfg RedisConfig, ctx *RedisContext) {
	infra.Redis[cfg] = ctx
}

func (infra *Infra) Start(ctx context.Context) error {
	dctx, cancel := context.WithTimeout(ctx, infra.StartupTimeout)
	defer cancel()

	type mongoResult struct {
		err      error
		cfg      MongoConfig
		mongoCtx MongoContext
	}

	type redisResult struct {
		err      error
		cfg      RedisConfig
		redisCtx RedisContext
	}

	mongoCh := make(chan mongoResult)
	redisCh := make(chan redisResult)

	wg := sync.WaitGroup{}
	done := make(chan struct{})

	for cfg := range infra.Mongo {
		wg.Go(func() {
			mongoCtx, err := connectToMongo(dctx, cfg)
			mongoCh <- mongoResult{cfg: cfg, mongoCtx: mongoCtx, err: err}
		})
	}
	for cfg := range infra.Redis {
		wg.Go(func() {
			redisCtx, err := ConnectToRedis(dctx, cfg)
			redisCh <- redisResult{cfg: cfg, redisCtx: redisCtx, err: err}
		})
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		case res := <-mongoCh:
			if res.err != nil {
				return res.err
			}
			mongoCtx := infra.Mongo[res.cfg]
			*mongoCtx = res.mongoCtx
		case res := <-redisCh:
			if res.err != nil {
				return res.err
			}
			redisCtx := infra.Redis[res.cfg]
			*redisCtx = res.redisCtx
		}
	}
}

func (infra *Infra) Shutdown(ctx context.Context) {
	dctx, cancel := context.WithTimeout(ctx, infra.ShutdownTimeout)
	defer cancel()

	wg := sync.WaitGroup{}
	done := make(chan struct{})
	errCh := make(chan error)

	for _, c := range infra.Mongo {
		wg.Go(func() {
			if err := disconnectFromMongo(dctx, *c); err != nil {
				errCh <- fmt.Errorf("'disconnectFromMongo' failed: %w", err)
			}
		})
	}
	for _, c := range infra.Redis {
		wg.Go(func() {
			if err := disconnectFromRedis(dctx, *c); err != nil {
				errCh <- fmt.Errorf("'disconnectFromRedis' failed: %w", err)
			}
		})
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case err := <-errCh:
			slog.Error("infrastructure shutdown failed", "error", err)
		}
	}
}
