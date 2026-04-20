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

	"github.com/AdventurerAmer/recipes-api/handlers"
	"github.com/AdventurerAmer/recipes-api/infra"
	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"
	ginRedis "github.com/gin-contrib/sessions/redis"
)

func main() {
	mongoCfg := infra.MongoConfig{
		Username: "admin",
		Password: "admin",
		Host:     "localhost",
		Port:     27017,
		Database: "dev",
	}
	redisCfg := infra.RedisConfig{
		Address:  "localhost:6379",
		Username: "",
		Password: "",
		Database: 0,
	}
	var mainDB infra.MongoContext
	var mainCache infra.RedisContext
	ifa := infra.New()
	ifa.BindMongo(mongoCfg, &mainDB)
	ifa.BindRedis(redisCfg, &mainCache)
	if err := ifa.Start(context.TODO()); err != nil {
		slog.Error("infrastructure startup failed", "error", err)
		os.Exit(1)
	}
	defer ifa.Shutdown(context.TODO())

	// TODO: hardcoding connections
	store, err := ginRedis.NewStore(10, "tcp", redisCfg.Address, redisCfg.Username, redisCfg.Password, []byte("xnx6D7fCxR47XqHGrnkqIBDjHIoz1csJ"))
	if err != nil {
		slog.Error("cache connection failed", "error", err)
		os.Exit(1)
	}

	recipesHandler := handlers.NewRecipesHandler(context.Background(), mainDB.Database.Collection("recipes"), mainCache.Client)
	authHandler := handlers.NewAuthHandler(context.Background(), mainDB.Client, mainDB.Database.Collection("users"))

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
