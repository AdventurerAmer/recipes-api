package infra

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Passward string `json:"password"`
	UseSSL   bool   `json:"userSSL"`
}

func (cfg *MinioConfig) Connect(ctx context.Context) (Disconnecter, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Username, cfg.Passward, ""),
		Secure: cfg.UseSSL,
	}
	type result struct {
		err    error
		client *minio.Client
	}
	ch := make(chan result)
	go func() {
		client, err := minio.New(cfg.Address, opts)
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
