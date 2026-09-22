package ports

import "context"

type TxnFunc func(ctx context.Context) error

type Transactor interface {
	WithTransaction(ctx context.Context, fn TxnFunc) error
}
