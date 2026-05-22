package ports

import "context"

type TextSearch interface {
	Index(ctx context.Context, index string, id string, doc any) error
	Search(ctx context.Context, index string, field string, value string, page, perPage int) ([][]byte, int, error)
	Delete(ctx context.Context, inded string, id string) error
}
