package infrastructure

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/AdventurerAmer/recipes-api/logging"
)

type Connecter interface {
	Connect(ctx context.Context) (Disconnecter, error)
}

type Disconnecter interface {
	Disconnect(ctx context.Context) error
}

type Infra struct {
	*Config
	logger     *logging.Logger
	components map[Connecter]Disconnecter
}

func New(logger *logging.Logger, opts ...Option) (*Infra, error) {
	cfg := &Config{
		startupTimeout:  2 * time.Second,
		shutdownTimeout: 4 * time.Second,
	}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	return &Infra{
		Config:     cfg,
		logger:     logger,
		components: make(map[Connecter]Disconnecter),
	}, nil
}

func (infra *Infra) Bind(connector Connecter, disconnector Disconnecter) {
	if _, ok := infra.components[connector]; ok {
		panic("connector is already bound")
	}
	infra.components[connector] = disconnector
}

func (infra *Infra) Start(ctx context.Context) error {
	dctx, cancel := context.WithTimeout(ctx, infra.startupTimeout)
	defer cancel()

	errCh := make(chan error)

	wg := sync.WaitGroup{}
	done := make(chan struct{})

	for conn, dstDisconn := range infra.components {
		wg.Go(func() {
			srcDisconn, err := conn.Connect(dctx)
			if err != nil {
				errCh <- err
			} else {
				ele := reflect.ValueOf(dstDisconn).Elem()
				ele.Set(reflect.ValueOf(srcDisconn).Elem())
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
			return ctx.Err()
		case <-done:
			return nil
		case err := <-errCh:
			return err
		}
	}
}

func (infra *Infra) Shutdown(ctx context.Context) {
	dctx, cancel := context.WithTimeout(ctx, infra.shutdownTimeout)
	defer cancel()

	wg := sync.WaitGroup{}
	done := make(chan struct{})
	errCh := make(chan error)

	for _, disconn := range infra.components {
		wg.Go(func() {
			if err := disconn.Disconnect(dctx); err != nil {
				errCh <- err
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

func (infra *Infra) Mongo(cfg *config.Mongo) *MongoContext {
	ctx := &MongoContext{}
	cfgWrapper := &Mongo{Mongo: cfg}
	infra.Bind(cfgWrapper, ctx)
	return ctx
}

func (infra *Infra) Redis(cfg *config.Redis) *RedisContext {
	ctx := &RedisContext{}
	cfgWrapper := &Redis{Redis: cfg}
	infra.Bind(cfgWrapper, ctx)
	return ctx
}

func (infra *Infra) RabbitMQ(cfg *config.RabbitMq) *RabbitMqContext {
	ctx := &RabbitMqContext{}
	cfgWrapper := &RabbitMq{RabbitMq: cfg, logger: infra.logger}
	infra.Bind(cfgWrapper, ctx)
	return ctx
}

func (infra *Infra) Minio(cfg *config.Minio) *MinioContext {
	ctx := &MinioContext{}
	cfgWrapper := &Minio{Minio: cfg}
	infra.Bind(cfgWrapper, ctx)
	return ctx
}

func (infra *Infra) ElasticSearch(cfg *config.ElasticSearch) *ElasticSearchContext {
	ctx := &ElasticSearchContext{}
	cfgWrapper := &ElasticSearch{ElasticSearch: cfg}
	infra.Bind(cfgWrapper, ctx)
	return ctx
}
