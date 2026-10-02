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
	Cache      ports.Cache
	TextSearch ports.TextSearch
	Transactor ports.Transactor
}

type mongoRepo struct {
	MongoConfig
	collection *mongo.Collection
}

func NewMongo(cfg MongoConfig) ports.RecipesRepository {
	return &mongoRepo{
		MongoConfig: cfg,
		collection:  cfg.Database.Collection("recipes"),
	}
}

func (repo *mongoRepo) Create(ctx context.Context, recipe *domain.Recipe) error {
	txn := func(tctx context.Context) error {
		result, err := repo.collection.InsertOne(tctx, recipe)
		if err != nil {
			return fmt.Errorf("'collection.InsertOne' failed: %w", mongoutils.ToDomainErr(err, "recipe"))
		}

		if err := repo.TextSearch.Index(tctx, "recipes", recipe.Id, recipe); err != nil {
			return fmt.Errorf("'textSearch.Index' failed: %w", err)
		}

		recipe.Id = result.InsertedID.(primitive.ObjectID).Hex()
		return nil
	}

	if err := repo.Transactor.WithTransaction(ctx, txn); err != nil {
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
	}

	filter := bson.M{"_id": oid}
	result := repo.collection.FindOne(ctx, filter)
	if err := result.Decode(&recipe); err != nil {
		return nil, fmt.Errorf("'result.Decode' failed: %w", mongoutils.ToDomainErr(err, "recipe"))
	}

	_ = repo.Cache.Put(ctx, key, recipe, 10*time.Minute)

	return &recipe, nil
}

func (repo *mongoRepo) List(ctx context.Context, cursor, userId string, limit int64) (*domain.Page[domain.Recipe], error) {
	key := composeRecipesCacheKey(cursor, userId, limit)
	var cachedPage domain.Page[domain.Recipe]
	if err := repo.Cache.Get(ctx, key, &cachedPage); err == nil {
		return &cachedPage, nil
	}

	filter := bson.M{}
	if userId != "" {
		filter["userId"] = userId
	}

	page, err := mongoutils.FindCursorPaginated[domain.Recipe](ctx, repo.collection, cursor, filter, limit)
	if err != nil {
		return nil, mongoutils.ToDomainErr(err, "recipe")
	}

	_ = repo.Cache.Put(ctx, key, page, 30*time.Second)

	return page, nil
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
			return nil, 0, fmt.Errorf("'json.Unmarshal' failed: %w", err)
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
		UpdatedAt    time.Time `bson:"updatedAt"`
		Version      int       `bson:"-"`
	}

	txn := func(tctx context.Context) error {
		filter := bson.M{"_id": oid, "version": recipe.Version}
		update := bson.D{
			{Key: "$set", Value: updateRecipeModel(*recipe)},
			{Key: "$inc", Value: bson.D{{Key: "version", Value: 1}}},
		}
		result, err := repo.collection.UpdateOne(tctx, filter, update)
		if err != nil {
			return fmt.Errorf("'collection.UpdateOne' failed: %w", mongoutils.ToDomainErr(err, "recipe"))
		}
		if result.MatchedCount == 0 {
			return errs.NewConflict(nil, "recipe was modified by another request")
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

	if err := repo.Transactor.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'txnMgr.WithTransaction' failed: %w", err)
	}

	return nil
}

func (repo *mongoRepo) Delete(ctx context.Context, recipe *domain.Recipe) error {
	oid, err := primitive.ObjectIDFromHex(recipe.Id)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}

	txn := func(tctx context.Context) error {
		filter := bson.M{"_id": oid, "version": recipe.Version}
		result, err := repo.collection.DeleteOne(tctx, filter)
		if err != nil {
			return fmt.Errorf("'collection.DeleteOne' failed: %w", err)
		}

		if result.DeletedCount == 0 {
			return errs.NewConflict(nil, "recipe was modified by another request")
		}

		if err := repo.Cache.Delete(ctx, composeRecipeCacheKey(recipe.Id)); err != nil {
			return fmt.Errorf("'Cache.Delet' failed: %w", err)
		}

		if err := repo.TextSearch.Delete(tctx, "recipes", recipe.Id); err != nil {
			return fmt.Errorf("'TextSearch.Delete' failed: %w", err)
		}
		return nil
	}
	if err := repo.Transactor.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'txnMgr.WithTransaction' failed: %w", err)
	}
	return nil
}
