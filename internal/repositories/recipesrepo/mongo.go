package recipesrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/errs"
	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/AdventurerAmer/recipes-api/mongoutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoConfig struct {
	Database   *mongo.Database
	Client     *mongo.Client
	Cache      ports.Cache
	TextSearch ports.TextSearch
}

type mongoRepo struct {
	MongoConfig
	collection *mongo.Collection
	txnMgr     *mongoutils.TxnManager
}

func NewMongo(cfg MongoConfig) ports.RecipesRepository {
	return &mongoRepo{
		MongoConfig: cfg,
		collection:  cfg.Database.Collection("recipes"),
		txnMgr:      mongoutils.NewTxnManager(cfg.Client),
	}
}

func (repo *mongoRepo) Create(ctx context.Context, recipe *domain.Recipe) error {
	txn := func(tctx context.Context) error {
		result, err := repo.collection.InsertOne(tctx, recipe)
		if err != nil {
			return fmt.Errorf("'collection.InsertOne' failed: %w", err)
		}

		if err := repo.TextSearch.Index(tctx, "recipes", recipe.Id, recipe); err != nil {
			return fmt.Errorf("'textSearch.Index' failed: %w", err)
		}

		recipe.Id = result.InsertedID.(primitive.ObjectID).Hex()
		return nil
	}

	if err := repo.txnMgr.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'txnMgr.WithTransaction' failed: %w", err)
	}

	return nil
}

func (repo *mongoRepo) Get(ctx context.Context, id string) (*domain.Recipe, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}
	var recipe domain.Recipe

	key := composeRecipeCacheKey(id)
	if err := repo.Cache.Get(ctx, key, &recipe); err == nil {
		return &recipe, nil
	} else if errs.IsNotFound(err) {
		defer func() {
			_ = repo.Cache.Put(ctx, key, recipe, 10*time.Second)
		}()
	}

	filter := bson.M{"_id": oid}
	result := repo.collection.FindOne(ctx, filter)
	if err := result.Decode(&recipe); err != nil {
		return nil, fmt.Errorf("'result.Decode' failed: %w", err)
	}
	return &recipe, nil
}

func (repo *mongoRepo) List(ctx context.Context, userId, lastId, sort string, limit int) ([]domain.Recipe, int, error) {
	filter := bson.M{}
	if lastId != "" {
		oid, err := primitive.ObjectIDFromHex(lastId)
		if err != nil {
			return nil, 0, fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
		}
		filter["_id"] = bson.M{"$gt": oid}
	}
	if userId != "" {
		filter["userId"] = userId
	}
	match := bson.D{{Key: "$match", Value: filter}}
	pagination := bson.D{
		{Key: "recipes", Value: bson.A{
			bson.D{{Key: "$sort", Value: mongoutils.ComposeSortStage(sort)}},
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
	defer mongoutils.CloseCursor(cursor)

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

func (repo *mongoRepo) Search(ctx context.Context, name string, page, pageSize int) ([]domain.Recipe, int, error) {
	results, total, err := repo.TextSearch.Search(ctx, "recipes", "name", name, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("'TextSearch.Search': %w", err)
	}

	recipes := make([]domain.Recipe, 0, len(results))
	for _, data := range results {
		var recipe domain.Recipe
		if err := json.Unmarshal(data, &recipe); err != nil {
			return nil, 0, err
		}
		recipes = append(recipes, recipe)
	}

	return recipes, total, nil
}

func (repo *mongoRepo) Update(ctx context.Context, recipe *domain.Recipe) error {
	oid, err := primitive.ObjectIDFromHex(recipe.Id)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}

	type updateRecipeModel struct {
		Id           string    `bson:"-"`
		CreatedAt    time.Time `bson:"-"`
		UserId       string    `bson:"-"`
		Name         string    `bson:"name"`
		Tags         []string  `bson:"tags"`
		Ingredients  []string  `bson:"ingredients"`
		Instructions []string  `bson:"instructions"`
		ImageURL     string    `bson:"imageURL"`
		Version      int       `bson:"-"`
	}

	txn := func(tctx context.Context) error {
		filter := bson.M{"_id": oid, "version": recipe.Version}
		update := bson.D{
			{Key: "$set", Value: updateRecipeModel(*recipe)},
			{Key: "$inc", Value: bson.D{{Key: "version", Value: 1}}},
		}
		if _, err := repo.collection.UpdateOne(tctx, filter, update); err != nil {
			return fmt.Errorf("'collection.UpdateOne' failed: %w", err)
		}

		if err := repo.Cache.Delete(ctx, composeRecipeCacheKey(recipe.Id)); err != nil {
			return fmt.Errorf("'Cache.Delet' failed: %w", err)
		}

		if err := repo.TextSearch.Index(tctx, "recipes", recipe.Id, recipe); err != nil {
			return fmt.Errorf("'textSearch.Index' failed: %w", err)
		}

		recipe.Version += 1
		return nil
	}

	if err := repo.txnMgr.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'txnMgr.WithTransaction' failed: %w", err)
	}

	return nil
}

func (repo *mongoRepo) Delete(ctx context.Context, userId, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}

	txn := func(tctx context.Context) error {
		filter := bson.M{"_id": oid, "userId": userId}
		if _, err := repo.collection.DeleteOne(tctx, filter); err != nil {
			return fmt.Errorf("'collection.DeleteOne' failed: %w", err)
		}

		if err := repo.Cache.Delete(ctx, composeRecipeCacheKey(id)); err != nil {
			return fmt.Errorf("'Cache.Delet' failed: %w", err)
		}

		if err := repo.TextSearch.Delete(tctx, "recipes", id); err != nil {
			return fmt.Errorf("'TextSearch.Delete' failed: %w", err)
		}
		return nil
	}
	if err := repo.txnMgr.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'txnMgr.WithTransaction' failed: %w", err)
	}
	return nil
}
