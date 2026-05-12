package ports

import (
	"context"
	"io"
)

const RecipeImagesBucketName = "recipes/images"

type ObjectStorageFile struct {
	Reader      io.Reader
	Size        int
	ContentType string
}

type ObjectStorage interface {
	GetURL(bucket, objectName string) string
	Upload(ctx context.Context, bucket, objectName string, file ObjectStorageFile) error
	Delete(ctx context.Context, bucket, objectName string) error
}
