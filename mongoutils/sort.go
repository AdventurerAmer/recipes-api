package mongoutils

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func ComposeSortStage(field string) bson.D {
	sortingOrder := 1
	if strings.HasPrefix(field, "-") {
		field, _ = strings.CutPrefix(field, "-")
		sortingOrder = -1
	}
	if field == "id" {
		field = "_id"
	}
	if field == "" {
		field = "createdAt"
	}
	sort := bson.D{{Key: field, Value: sortingOrder}}
	if field != "_id" {
		sort = append(sort, bson.E{Key: "_id", Value: -1})
	}
	return sort
}
