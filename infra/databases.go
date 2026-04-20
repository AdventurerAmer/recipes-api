package infra

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
}

type MongoContext struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func connectToMongo(ctx context.Context, cfg MongoConfig) (MongoContext, error) {
	// TODO: distingus between 'dev' and 'prod' in terms of authentication
	connStr := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	opts := options.Client().ApplyURI(connStr)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return MongoContext{}, fmt.Errorf("'mongo.Connect' failed: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return MongoContext{}, fmt.Errorf("'client.Ping' failed: %w", err)
	}

	db := client.Database(cfg.Database)
	return MongoContext{
		Client:   client,
		Database: db,
	}, nil
}

func disconnectFromMongo(ctx context.Context, mongoCtx MongoContext) error {
	if err := mongoCtx.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("'Client.Disconnect' failed: %w", err)
	}
	return nil
}
