package infra

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	*config.Minio
}

func (cfg *Minio) Connect(ctx context.Context) (Disconnecter, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Username, cfg.Passward, ""),
		Secure: cfg.UseTLS,
	}
	type result struct {
		err    error
		client *minio.Client
	}
	ch := make(chan result)
	go func() {
		client, err := minio.New(cfg.Addr(), opts)
		if err != nil {
			err = fmt.Errorf("'minio.New' failed: %w", err)
		}
		ch <- result{client: client, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return nil, res.err
		}
		return &MinioContext{client: res.client}, nil
	}
}

type MinioContext struct {
	client *minio.Client
}

func (c *MinioContext) Disconnect(context.Context) error {
	return nil
}
