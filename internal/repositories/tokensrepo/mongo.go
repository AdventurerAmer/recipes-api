package tokensrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/errs"
	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/AdventurerAmer/recipes-api/mongoutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	Database   *mongo.Database
	Transactor ports.Transactor
	Cache      ports.Cache
}

type mongoRepo struct {
	MongoConfig
	collection *mongo.Collection
}

func NewMongo(cfg MongoConfig) ports.TokensRepository {
	return &mongoRepo{
		MongoConfig: cfg,
		collection:  cfg.Database.Collection("tokens"),
	}
}

func (repo *mongoRepo) Create(ctx context.Context, token *domain.Token) error {
	result, err := repo.collection.InsertOne(ctx, token)
	if err != nil {
		return fmt.Errorf("'collection.InsertOne' failed: %w", err)
	}
	token.Id = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (repo *mongoRepo) Get(ctx context.Context, tokenType domain.TokenType, hash string) (*domain.Token, error) {
	var token domain.Token

	key := composeCacheKey(tokenType, hash)
	if err := repo.Cache.Get(ctx, key, &token); err == nil {
		return &token, nil
	}

	filter := bson.M{"type": tokenType, "hash": hash, "usedAt": nil}
	opts := options.FindOne().SetSort(bson.M{"createdAt": -1})
	result := repo.collection.FindOne(ctx, filter, opts)
	if err := result.Decode(&token); err != nil {
		return nil, fmt.Errorf("'collection.FindOne' failed: %w", mongoutils.ToDomainErr(err, "token"))
	}

	_ = repo.Cache.Put(ctx, key, token, 10*time.Minute)

	return &token, nil
}

func (repo *mongoRepo) Update(ctx context.Context, token *domain.Token) error {
	oid, err := primitive.ObjectIDFromHex(token.Id)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}

	type updateModel struct {
		Id        string           `bson:"-"`
		CreatedAt time.Time        `bson:"-"`
		Type      domain.TokenType `bson:"-"`
		UserId    string           `bson:"-"`
		Hash      string           `bson:"-"`
		ExpiresAt time.Time        `bson:"-"`
		UsedAt    *time.Time       `bson:"usedAt"`
	}

	txn := func(tctx context.Context) error {
		filter := bson.M{"_id": oid}
		update := bson.D{
			{Key: "$set", Value: updateModel(*token)},
			{Key: "$inc", Value: bson.D{{Key: "version", Value: 1}}},
		}
		result, err := repo.collection.UpdateOne(tctx, filter, update)
		if err != nil {
			return fmt.Errorf("'collection.UpdateOne' failed: %w", mongoutils.ToDomainErr(err, "user"))
		}
		if result.MatchedCount == 0 {
			return errs.NewConflict(nil, "user was modified by another request")
		}

		key := composeCacheKey(token.Type, token.Hash)
		if err := repo.Cache.Delete(ctx, key); err != nil {
			return fmt.Errorf("'Cache.Delete' failed: %w", err)
		}

		return nil
	}
	if err := repo.Transactor.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'txnMgr.WithTransaction' failed: %w", err)
	}

	return nil
}

func (repo *mongoRepo) Delete(ctx context.Context, token *domain.Token) error {
	oid, err := primitive.ObjectIDFromHex(token.Id)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}

	txn := func(tctx context.Context) error {
		filter := bson.M{"_id": oid}
		result, err := repo.collection.DeleteOne(ctx, filter)
		if err != nil {
			return fmt.Errorf("'collection.DeleteOne' failed: %w", err)
		}
		if result.DeletedCount == 0 {
			return errs.NewConflict(nil, "user was modified by another request")
		}

		key := composeCacheKey(token.Type, token.Hash)
		if err := repo.Cache.Delete(ctx, key); err != nil {
			return fmt.Errorf("'Cache.Delete' failed: %w", err)
		}

		return nil
	}
	if err := repo.Transactor.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'Transactor.WithTransaction' failed: %w", err)
	}

	return nil
}
