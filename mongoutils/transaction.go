package mongoutils

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type txnManager struct {
	client *mongo.Client
}

func NewTransactor(client *mongo.Client) ports.Transactor {
	return &txnManager{client: client}
}

func (tm *txnManager) WithTransaction(ctx context.Context, fn ports.TxnFunc) error {
	if mongo.SessionFromContext(ctx) != nil {
		return fn(ctx)
	}

	session, err := tm.client.StartSession()
	if err != nil {
		return fmt.Errorf("'client.StartSession' failed: %w", err)
	}
	defer session.EndSession(ctx)

	wc := writeconcern.Majority()
	rc := readconcern.Snapshot()
	opts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	callback := func(sctx mongo.SessionContext) (any, error) {
		if err := fn(sctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if _, err := session.WithTransaction(ctx, callback, opts); err != nil {
		return fmt.Errorf("'session.WithTransaction' failed: %w", err)
	}

	return nil
}
