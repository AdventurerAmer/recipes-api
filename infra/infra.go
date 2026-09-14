package infra

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/AdventurerAmer/recipes-api/config"
)

type Connecter interface {
	Connect(ctx context.Context) (Disconnecter, error)
}

type Disconnecter interface {
	Disconnect(ctx context.Context) error
}

type Infra struct {
	StartupTimeout  time.Duration
	ShutdownTimeout time.Duration
	Components      map[Connecter]Disconnecter
}

func New() *Infra {
	return &Infra{
		StartupTimeout:  time.Second,
		ShutdownTimeout: time.Second,
		Components:      make(map[Connecter]Disconnecter),
	}
}

func (infra *Infra) Bind(connector Connecter, disconnector Disconnecter) {
	if _, ok := infra.Components[connector]; ok {
		panic("connector is already bound")
	}
	infra.Components[connector] = disconnector
}

func (infra *Infra) Start(ctx context.Context) error {
	dctx, cancel := context.WithTimeout(ctx, infra.StartupTimeout)
	defer cancel()

	errCh := make(chan error)

	wg := sync.WaitGroup{}
	done := make(chan struct{})

	for conn, dstDisconn := range infra.Components {
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
	dctx, cancel := context.WithTimeout(ctx, infra.ShutdownTimeout)
	defer cancel()

	wg := sync.WaitGroup{}
	done := make(chan struct{})
	errCh := make(chan error)

	for _, disconn := range infra.Components {
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

func (infra *Infra) BindMongo(cfg *config.Mongo, ctx *MongoContext) {
	wrapper := &Mongo{Mongo: cfg}
	infra.Bind(wrapper, ctx)
}

func (infra *Infra) BindRedis(cfg *config.Redis, ctx *RedisContext) {
	wrapper := &Redis{Redis: cfg}
	infra.Bind(wrapper, ctx)
}

func (infra *Infra) BindMinio(cfg *config.Minio, ctx *MinioContext) {
	wrapper := &Minio{Minio: cfg}
	infra.Bind(wrapper, ctx)
}

func (infra *Infra) BindElasticSearch(cfg *config.ElasticSearch, ctx *ElasticSearchContext) {
	wrapper := &ElasticSearch{ElasticSearch: cfg}
	infra.Bind(wrapper, ctx)
}
