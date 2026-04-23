package recipesrepo

import (
	"context"
	"fmt"
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

func (repo *mongoRepo) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}
	filter := bson.M{"_id": oid}
	if _, err := repo.collection.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("'collection.DeleteOne' failed: %w", err)
	}
	return nil
}
