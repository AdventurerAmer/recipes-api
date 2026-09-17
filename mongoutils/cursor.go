package mongoutils

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func CloseCursor(cursor *mongo.Cursor) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// TODO: add retries here and make it async.
	_ = cursor.Close(ctx)
}
