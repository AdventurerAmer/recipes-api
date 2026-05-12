package minio

import (
	"context"
	"fmt"

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

func (mos *minioObjectStorage) GetURL(bucket, objectName string) string {
	endpoint := mos.client.EndpointURL()
	return fmt.Sprintf("http://%s/%s/%s", endpoint, bucket, objectName)
}

func (mos *minioObjectStorage) Upload(ctx context.Context, bucket, objectName string, file ports.ObjectStorageFile) error {
	opts := minio.PutObjectOptions{
		ContentType: file.ContentType,
	}
	_, err := mos.client.PutObject(ctx, bucket, objectName, file.Reader, int64(file.Size), opts)
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
