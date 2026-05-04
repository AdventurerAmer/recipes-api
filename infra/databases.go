package infra

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	Username string `cfg:"username"`
	Password string `cfg:"password"`
	Host     string `cfg:"host"`
	Port     int    `cfg:"port"`
	Name     string `cfg:"name"`
}

type MongoContext struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func connectToMongo(ctx context.Context, cfg MongoConfig) (MongoContext, error) {
	// TODO: distingus between 'dev' and 'prod' in terms of authentication
	connStr := fmt.Sprintf("mongodb://%s:%d/%s?replicaSet=rs0", cfg.Host, cfg.Port, cfg.Name)
	opts := options.Client().ApplyURI(connStr)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return MongoContext{}, fmt.Errorf("'mongo.Connect' failed: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return MongoContext{}, fmt.Errorf("'client.Ping' failed: %w", err)
	}

	db := client.Database(cfg.Name)
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
