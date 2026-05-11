package ports

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int, contentType string) error
	Delete(ctx context.Context, bucket, objectName string) error
}
