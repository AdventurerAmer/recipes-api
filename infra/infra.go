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
	Minio           map[MinioConfig]*MinioContext
}

func New() *Infra {
	return &Infra{
		StartupTimeout:  time.Second,
		ShutdownTimeout: time.Second,
		Mongo:           make(map[MongoConfig]*MongoContext),
		Redis:           make(map[RedisConfig]*RedisContext),
		Minio:           make(map[MinioConfig]*MinioContext),
	}
}

func (infra *Infra) BindMongo(cfg MongoConfig, ctx *MongoContext) {
	if _, ok := infra.Mongo[cfg]; ok {
		panic("cfg is already bound")
	}
	infra.Mongo[cfg] = ctx
}

func (infra *Infra) BindRedis(cfg RedisConfig, ctx *RedisContext) {
	if _, ok := infra.Redis[cfg]; ok {
		panic("cfg is already bound")
	}
	infra.Redis[cfg] = ctx
}

func (infra *Infra) BindMinio(cfg MinioConfig, ctx *MinioContext) {
	if _, ok := infra.Minio[cfg]; ok {
		panic("cfg is already bound")
	}
	infra.Minio[cfg] = ctx
}

func (infra *Infra) Start(ctx context.Context) error {
	dctx, cancel := context.WithTimeout(ctx, infra.StartupTimeout)
	defer cancel()

	type mongoResult struct {
		err error
		cfg MongoConfig
		ctx MongoContext
	}

	type redisResult struct {
		err error
		cfg RedisConfig
		ctx RedisContext
	}

	type minioResult struct {
		err error
		cfg MinioConfig
		ctx MinioContext
	}

	mongoCh := make(chan mongoResult)
	redisCh := make(chan redisResult)
	minioCh := make(chan minioResult)

	wg := sync.WaitGroup{}
	done := make(chan struct{})

	for cfg := range infra.Mongo {
		wg.Go(func() {
			mongoCtx, err := connectToMongo(dctx, cfg)
			mongoCh <- mongoResult{cfg: cfg, ctx: mongoCtx, err: err}
		})
	}
	for cfg := range infra.Redis {
		wg.Go(func() {
			redisCtx, err := connectToRedis(dctx, cfg)
			redisCh <- redisResult{cfg: cfg, ctx: redisCtx, err: err}
		})
	}

	for cfg := range infra.Minio {
		wg.Go(func() {
			minioCtx, err := connectToMinio(dctx, cfg)
			minioCh <- minioResult{cfg: cfg, ctx: minioCtx, err: err}
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
			*mongoCtx = res.ctx
		case res := <-redisCh:
			if res.err != nil {
				return res.err
			}
			redisCtx := infra.Redis[res.cfg]
			*redisCtx = res.ctx
		case res := <-minioCh:
			if res.err != nil {
				return res.err
			}
			minioCtx := infra.Minio[res.cfg]
			*minioCtx = res.ctx
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
	for _, c := range infra.Minio {
		wg.Go(func() {
			if err := disconnectFromMinio(dctx, *c); err != nil {
				errCh <- fmt.Errorf("'disconnectFromMinio' failed: %w", err)
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
