package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/minio/minio-go/v7"
)

type minioObjectStorage struct {
	client *minio.Client
}

func New(client *minio.Client) ports.ObjectStorage {
	return &minioObjectStorage{
		client: client,
	}
}

func (mos *minioObjectStorage) Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int, contentType string) error {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}
	_, err := mos.client.PutObject(ctx, bucket, objectName, reader, int64(size), opts)
	if err != nil {
		return fmt.Errorf("'client.PutObject' failed: %w", err)
	}
	return nil
}

func (mos *minioObjectStorage) Delete(ctx context.Context, bucket, objectName string) error {
	opts := minio.RemoveObjectOptions{}
	err := mos.client.RemoveObject(ctx, bucket, objectName, opts)
	if err != nil {
		return fmt.Errorf("'client.RemoveObject' failed: %w", err)
	}
	return nil
}
