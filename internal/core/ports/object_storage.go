package ports

import (
	"context"
	"io"
)

type BucketName string

const (
	BucketNameImages BucketName = "images"
)

type ObjectStorageFile struct {
	Reader      io.Reader
	Size        int
	ContentType string
}

type ObjectStorage interface {
	GetURL(bucket BucketName, name string) string
	Upload(ctx context.Context, bucket BucketName, name string, file ObjectStorageFile) error
	Delete(ctx context.Context, bucket BucketName, name string) error
}
