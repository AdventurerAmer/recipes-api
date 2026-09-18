package mongoutils

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type TxnFunc func(ctx context.Context) error

type TxnManager struct {
	client *mongo.Client
}

func NewTxnManager(client *mongo.Client) *TxnManager {
	return &TxnManager{client: client}
}

func (tm *TxnManager) WithTransaction(ctx context.Context, fn TxnFunc) error {
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
