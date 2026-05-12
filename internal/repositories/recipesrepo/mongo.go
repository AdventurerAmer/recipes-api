package recipesrepo

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoConfig struct {
	Database *mongo.Database
}

type mongoRepo struct {
	collection *mongo.Collection
}

func NewMongo(cfg MongoConfig) ports.RecipesRepository {
	return &mongoRepo{
		collection: cfg.Database.Collection("recipes"),
	}
}

func (repo *mongoRepo) Create(ctx context.Context, recipe *domain.Recipe) error {
	result, err := repo.collection.InsertOne(ctx, recipe)
	if err != nil {
		return fmt.Errorf("'collection.InsertOne' failed: %w", err)
	}
	recipe.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (repo *mongoRepo) Get(ctx context.Context, id string) (domain.Recipe, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.Recipe{}, fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}
	filter := bson.M{"_id": oid}
	result := repo.collection.FindOne(ctx, filter)
	var recipe domain.Recipe
	if err := result.Decode(&recipe); err != nil {
		return domain.Recipe{}, fmt.Errorf("'result.Decode' failed: %w", err)
	}
	return recipe, nil
}

func (repo *mongoRepo) List(ctx context.Context, lastID, userID, sortBy string, limit int) ([]domain.Recipe, int, error) {
	filter := bson.M{}
	if lastID != "" {
		oid, err := primitive.ObjectIDFromHex(lastID)
		if err != nil {
			return nil, 0, fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
		}
		filter["_id"] = bson.M{"$gt": oid}
	}
	if userID != "" {
		filter["userID"] = userID
	}
	match := bson.D{{Key: "$match", Value: filter}}
	sortingOrder := 1
	if strings.HasPrefix(sortBy, "-") {
		sortBy, _ = strings.CutPrefix(sortBy, "-")
		sortingOrder = -1
	}
	if sortBy == "id" {
		sortBy = "_id"
	}
	if sortBy == "" {
		sortBy = "createdAt"
	}
	sort := bson.D{{Key: sortBy, Value: sortingOrder}}
	if sortBy != "_id" {
		sort = append(sort, bson.E{Key: "_id", Value: -1})
	}

	pagination := bson.D{
		{Key: "recipes", Value: bson.A{
			bson.D{{Key: "$sort", Value: sort}},
			bson.D{{Key: "$limit", Value: limit}},
		}},
		{Key: "total", Value: bson.A{
			bson.D{{Key: "$count", Value: "count"}},
		}},
	}
	facet := bson.D{{Key: "$facet", Value: pagination}}
	pipeline := mongo.Pipeline{match, facet}
	cursor, err := repo.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, fmt.Errorf("'collection.Aggregate' failed: %w", err)
	}
	defer func() {
		cctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := cursor.Close(cctx); err != nil {
			slog.Error("'cursor.Close' failed", "error", err)
		}
	}()

	type Total struct {
		Count int `bson:"count"`
	}

	var results []struct {
		Recipes []domain.Recipe `bson:"recipes"`
		Totals  []Total         `bson:"total"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, fmt.Errorf("'cursor.All' failed: %w", err)
	}
	if len(results) == 0 {
		return nil, 0, nil
	}
	result := results[0]
	total := 0
	if len(result.Totals) != 0 {
		total = result.Totals[0].Count
	}
	return result.Recipes, total, nil
}

func (repo *mongoRepo) Update(ctx context.Context, recipe *domain.Recipe) error {
	oid, err := primitive.ObjectIDFromHex(recipe.ID)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}
	type updateRecipeModel struct {
		ID           string    `bson:"-"`
		CreatedAt    time.Time `bson:"-"`
		UserID       string    `bson:"-"`
		Name         string    `bson:"name"`
		Tags         []string  `bson:"tags"`
		Ingredients  []string  `bson:"ingredients"`
		Instructions []string  `bson:"instructions"`
		Image        string    `bson:"image"`
		Version      int       `bson:"-"`
	}
	filter := bson.M{"_id": oid, "version": recipe.Version}
	update := bson.D{
		{Key: "$set", Value: updateRecipeModel(*recipe)},
		{
			Key: "$inc", Value: bson.D{
				{Key: "version", Value: 1},
			},
		},
	}
	if _, err := repo.collection.UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("'collection.UpdateOne' failed: %w", err)
	}
	recipe.Version += 1
	return nil
}

func (repo *mongoRepo) Delete(ctx context.Context, userID, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}
	filter := bson.M{"_id": oid, "userID": userID}
	if _, err := repo.collection.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("'collection.DeleteOne' failed: %w", err)
	}
	return nil
}
