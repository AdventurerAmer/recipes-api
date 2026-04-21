package usersrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoConfig struct {
	database *mongo.Database
	client   *mongo.Client
}

type mongoRepo struct {
	MongoConfig
	collection *mongo.Collection
}

func NewMongo(cfg MongoConfig) ports.UsersRepository {
	return &mongoRepo{
		MongoConfig: cfg,
		collection:  cfg.database.Collection("users"),
	}
}

func (repo *mongoRepo) Create(ctx context.Context, user *domain.User) error {
	session, err := repo.client.StartSession()
	if err != nil {
		return fmt.Errorf("'client.StartSession' failed: %w", err)
	}
	defer session.EndSession(ctx)

	txn := func(ctx mongo.SessionContext) (any, error) {
		filter := bson.M{"username": user.Username}
		findResult := repo.collection.FindOne(ctx, filter)
		var findUser domain.User
		if err := findResult.Decode(&findUser); err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, fmt.Errorf("'collection.FindOne' failed: %w", err)
			}
		}
		result, err := repo.collection.InsertOne(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("'collection.InsertOne' failed: %w", err)
		}
		user.ID = result.InsertedID.(primitive.ObjectID).Hex()
		return nil, nil
	}
	if _, err := session.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'session.WithTransaction' failed: %w", err)
	}

	return nil
}

func (repo *mongoRepo) Get(ctx context.Context, id string) (domain.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.User{}, fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}
	filter := bson.M{"_id": oid}
	result := repo.collection.FindOne(ctx, filter)
	var user domain.User
	if err := result.Decode(&user); err != nil {
		return domain.User{}, fmt.Errorf("'collection.FindOne' failed: %w", err)
	}
	return user, nil
}

func (repo *mongoRepo) GetByName(ctx context.Context, username string) (domain.User, error) {
	filter := bson.M{"username": username}
	result := repo.collection.FindOne(ctx, filter)
	var user domain.User
	if err := result.Decode(&user); err != nil {
		return domain.User{}, fmt.Errorf("'collection.FindOne' failed: %w", err)
	}
	return user, nil
}

func (repo *mongoRepo) Update(ctx context.Context, user *domain.User) error {
	oid, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}
	filter := bson.M{"_id": oid}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "password", Value: user.Password}}},
		{Key: "$inc", Value: bson.D{{Key: "version", Value: 1}}},
	}
	if _, err := repo.collection.UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("'collection.UpdateOne' failed: %w", err)
	}
	user.Version += 1
	return nil
}

func (repo *mongoRepo) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}
	filter := bson.M{"_id": oid}
	if _, err := repo.collection.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("'collection.UpdateOne' failed: %w", err)
	}
	return nil
}
