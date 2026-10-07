package mongoutils

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func FindCursorPaginated[T domain.Pager](ctx context.Context,
	collection *mongo.Collection,
	cursorStr string,
	filter bson.M,
	limit int64,
) (*domain.Page[T], error) {
	sort := bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}
	opts := options.Find().SetSort(sort).SetLimit(limit + 1)

	query := bson.M{}
	maps.Copy(query, filter)

	if cursorStr != "" {
		cursor, err := domain.DecodeCursor(cursorStr)
		if err != nil {
			return nil, fmt.Errorf("'domain.DecodeCursor' failed: %w", err)
		}
		query["$or"] = []bson.M{
			{"createdAt": bson.M{"$lt": cursor.CreatedAt}},
			{"createdAt": cursor.CreatedAt, "_id": bson.M{"$lt": cursor.Id}},
		}
	}

	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("'collection.Find' failed: %w", err)
	}
	defer func() {
		go closeCursor(ctx, cursor)
	}()

	var items []T
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("'cursor.All' failed: %w", err)
	}

	hasNext := int64(len(items)) > limit
	if hasNext {
		items = items[:limit]
	}

	page := &domain.Page[T]{
		Items:   items,
		HasNext: hasNext,
	}

	if len(page.Items) > 0 {
		last := page.Items[len(page.Items)-1]
		nextCursor := domain.Cursor{
			Id:        last.GetId(),
			CreatedAt: last.GetCreatedAt(),
		}
		encoded, err := nextCursor.Encode()
		if err != nil {
			return nil, fmt.Errorf("'nextCursor.Encode' failed: %w", err)
		}
		page.NextCursor = encoded
	}

	return page, nil
}

func closeCursor(parent context.Context, cursor *mongo.Cursor) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := cursor.Close(ctx); err != nil {
		logger := logging.Get(parent)
		logger.Error("close cursor failed", "error", err)
	}
}
