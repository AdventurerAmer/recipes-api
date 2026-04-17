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

	"github.com/gin-contrib/sessions"
	ginRedis "github.com/gin-contrib/sessions/redis"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mongoOpts := options.Client().ApplyURI("mongodb://admin:admin@localhost:27017/dev?authSource=admin")
	mongoClient, err := mongo.Connect(ctx, mongoOpts)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
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

	db := mongoClient.Database("dev")

	store, err := ginRedis.NewStore(10, "tcp", "localhost:6379", "", "", []byte("xnx6D7fCxR47XqHGrnkqIBDjHIoz1csJ"))
	if err != nil {
		slog.Error("cache connection failed", "error", err)
		os.Exit(1)
	}

	recipesHandler := handlers.NewRecipesHandler(context.Background(), db.Collection("recipes"), redisClient)
	authHandler := handlers.NewAuthHandler(context.Background(), mongoClient, db.Collection("users"))

	r := gin.Default()
	r.POST("/signup", authHandler.SignUpHandler)
	r.POST("/signIn", authHandler.SignInHandler)
	r.POST("/signout", authHandler.SignOutHandler)

	r.GET("/recipes", recipesHandler.ListRecipesHandler)
	r.GET("/recipes/search", recipesHandler.SearchRecipesHandler)

	authed := r.Group("/")
	authed.Use(sessions.Sessions("recipes", store))
	authed.Use(authHandler.AuthMiddleware())
	{
		authed.POST("/recipes", recipesHandler.NewRecipeHandler)
		authed.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
		authed.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	}
	r.Run(":3000")
}
