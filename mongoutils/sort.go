package mongoutils

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func ComposeSortStage(field string) bson.D {
	order := 1
	if strings.HasPrefix(field, "-") {
		field, _ = strings.CutPrefix(field, "-")
		order = -1
	}
	if field == "id" {
		field = "_id"
	}
	if field == "" {
		field = "createdAt"
	}
	stage := bson.D{{Key: field, Value: order}}
	if field != "_id" {
		stage = append(stage, bson.E{Key: "_id", Value: -1})
	}
	return stage
}
