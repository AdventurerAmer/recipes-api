package infra

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	*config.Mongo
}

func (cfg *Mongo) Connect(ctx context.Context) (Disconnecter, error) {
	// TODO: distingus between 'dev' and 'prod' in terms of authentication
	connStr := fmt.Sprintf("mongodb://%s:%d/%s?replicaSet=rs0", cfg.Host, cfg.Port, cfg.Name)
	opts := options.Client().ApplyURI(connStr)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("'mongo.Connect' failed: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("'client.Ping' failed: %w", err)
	}

	return &MongoContext{
		Client:   client,
		Database: client.Database(cfg.Name),
	}, nil
}

type MongoContext struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func (c *MongoContext) Disconnect(ctx context.Context) error {
	if err := c.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("'Client.Disconnect' failed: %w", err)
	}
	c.Client = nil
	c.Database = nil
	return nil
}
