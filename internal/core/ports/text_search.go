package ports

import (
	"context"
	"encoding/json"
)

type TextSearch interface {
	Index(ctx context.Context, index string, id string, doc any) error
	Search(ctx context.Context, index, cursor string, limit int64) ([]json.RawMessage, string, error)
	Delete(ctx context.Context, inded string, id string) error
}
