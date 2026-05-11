package infra

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Addr     string `json:"addr"`
	Username string `json:"username"`
	Passward string `json:"password"`
	UseSSL   bool   `json:"userSSL"`
}

type MinioContext struct {
	client *minio.Client
}

func connectToMinio(ctx context.Context, cfg MinioConfig) (MinioContext, error) {
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
		client, err := minio.New(cfg.Addr, opts)
		if err != nil {
			err = fmt.Errorf("'minio.New' failed: %w", err)
		}
		ch <- result{client: client, err: err}
	}()
	select {
	case <-ctx.Done():
		return MinioContext{}, ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return MinioContext{}, res.err
		}
		return MinioContext{client: res.client}, nil
	}
}

func disconnectFromMinio(_ context.Context, _ MinioContext) error {
	return nil
}
