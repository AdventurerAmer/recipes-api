package usersrepo

import (
	"context"
	"errors"
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
	Database *mongo.Database
	Client   *mongo.Client
	Cache    ports.Cache
}

type mongoRepo struct {
	MongoConfig
	collection *mongo.Collection
	txnMgr     *mongoutils.TxnManager
}

func NewMongo(cfg MongoConfig) ports.UsersRepository {
	return &mongoRepo{
		MongoConfig: cfg,
		collection:  cfg.Database.Collection("users"),
		txnMgr:      mongoutils.NewTxnManager(cfg.Client),
	}
}

func (repo *mongoRepo) Create(ctx context.Context, user *domain.User) error {
	txn := func(tctx context.Context) error {
		var findUser domain.User
		filter := bson.M{"email": user.Email}
		findResult := repo.collection.FindOne(tctx, filter)
		if err := findResult.Decode(&findUser); err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return fmt.Errorf("'collection.FindOne' failed: %w", err)
			}
		} else {
			return errs.NewAlreadyExists(err, "user already exists")
		}

		result, err := repo.collection.InsertOne(tctx, user)
		if err != nil {
			return fmt.Errorf("'collection.InsertOne' failed: %w", err)
		}
		user.Id = result.InsertedID.(primitive.ObjectID).Hex()

		return nil
	}
	if err := repo.txnMgr.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'txnMgr.WithTransaction' failed: %w", err)
	}

	return nil
}

func (repo *mongoRepo) GetById(ctx context.Context, id string) (*domain.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}

	var user domain.User

	key := composeUserByIdCacheKey(id)
	if err := repo.Cache.Get(ctx, key, &user); err == nil {
		return &user, nil
	} else if errs.IsNotFound(err) {
		defer func() {
			_ = repo.Cache.Put(ctx, key, user, 10*time.Second)
		}()
	}

	filter := bson.M{"_id": oid}
	result := repo.collection.FindOne(ctx, filter)
	if err := result.Decode(&user); err != nil {
		return nil, fmt.Errorf("'collection.FindOne' failed: %w", err)
	}

	return &user, nil
}

func (repo *mongoRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	key := composeUserByEmailCacheKey(email)
	if err := repo.Cache.Get(ctx, key, &user); err == nil {
		return &user, nil
	} else if errs.IsNotFound(err) {
		defer func() {
			_ = repo.Cache.Put(ctx, key, user, 10*time.Second)
		}()
	}

	filter := bson.M{"email": email}
	result := repo.collection.FindOne(ctx, filter)
	if err := result.Decode(&user); err != nil {
		return nil, fmt.Errorf("'collection.FindOne' failed: %w", err)
	}
	return &user, nil
}

func (repo *mongoRepo) Update(ctx context.Context, user *domain.User) error {
	oid, err := primitive.ObjectIDFromHex(user.Id)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}

	type userUpdateModel struct {
		Id           string    `bson:"-"`
		CreatedAt    time.Time `bson:"-"`
		Email        string    `bson:"email"`
		DisplayName  string    `bson:"displayName"`
		PasswordHash string    `bson:"passwordHash"`
		Version      int       `bson:"-"`
	}

	txn := func(tctx context.Context) error {
		filter := bson.M{"_id": oid}
		update := bson.D{
			{Key: "$set", Value: userUpdateModel(*user)},
			{Key: "$inc", Value: bson.D{{Key: "version", Value: 1}}},
		}
		if _, err := repo.collection.UpdateOne(tctx, filter, update); err != nil {
			return fmt.Errorf("'collection.UpdateOne' failed: %w", err)
		}

		keys := []string{
			composeUserByIdCacheKey(user.Id),
			composeUserByEmailCacheKey(user.Email),
		}
		if err := repo.Cache.Delete(ctx, keys...); err != nil {
			return fmt.Errorf("'Cache.Delete' failed: %w", err)
		}

		user.Version += 1

		return nil
	}
	if err := repo.txnMgr.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'txnMgr.WithTransaction' failed: %w", err)
	}

	return nil
}

func (repo *mongoRepo) Delete(ctx context.Context, user *domain.User) error {
	oid, err := primitive.ObjectIDFromHex(user.Id)
	if err != nil {
		return fmt.Errorf("'primitive.ObjectIDFromHex' failed: %w", err)
	}

	txn := func(tctx context.Context) error {
		filter := bson.M{"_id": oid}
		if _, err := repo.collection.DeleteOne(ctx, filter); err != nil {
			return fmt.Errorf("'collection.DeleteOne' failed: %w", err)
		}

		keys := []string{
			composeUserByIdCacheKey(user.Id),
			composeUserByEmailCacheKey(user.Email),
		}
		if err := repo.Cache.Delete(ctx, keys...); err != nil {
			return fmt.Errorf("'Cache.Delete' failed: %w", err)
		}

		return nil
	}
	if err := repo.txnMgr.WithTransaction(ctx, txn); err != nil {
		return fmt.Errorf("'txnMgr.WithTransaction' failed: %w", err)
	}

	return nil
}
