// Recipes API
//
// This is a sample recipes API.
//
//	Schemes: http
//	Host: localhost:3000
//	BasePath: /
//	Version: 1.0.0
//	Contact: Ahmed Amer
//
// <ahamerdev@gmail.com>
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/AdventurerAmer/recipes-api/handlers"
	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mongoOpts := options.Client().ApplyURI("mongodb://admin:admin@localhost:27017/dev?authSource=admin")
	mongoCilent, err := mongo.Connect(ctx, mongoOpts)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	if err := mongoCilent.Ping(ctx, readpref.Primary()); err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	redisOpts := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	redisClient := redis.NewClient(redisOpts)

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		slog.Error("cache connection failed", "error", err)
		os.Exit(1)
	}

	collection := mongoCilent.Database("dev").Collection("recipes")

	handlers := handlers.NewRecipesHandler(context.Background(), collection, redisClient)
	r := gin.Default()
	r.POST("/recipes", handlers.NewRecipeHandler)
	r.GET("/recipes", handlers.ListRecipesHandler)
	r.GET("/recipes/search", handlers.SearchRecipesHandler)
	r.PUT("/recipes/:id", handlers.UpdateRecipeHandler)
	r.DELETE("/recipes/:id", handlers.DeleteRecipeHandler)
	r.Run(":3000")
}
