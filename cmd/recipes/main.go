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
	"net/http"
	"os"

	"github.com/AdventurerAmer/recipes-api/handlers"
	"github.com/AdventurerAmer/recipes-api/infra"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/recipessrv"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/userssrv"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/recipesrepo"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/usersrepo"
	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"
	ginRedis "github.com/gin-contrib/sessions/redis"
)

type App struct {
	mainDB    infra.MongoContext
	mainCache infra.RedisContext
}

func main() {
	app := &App{}
	infraCtx := infra.New()
	infraCtx.BindMongo(infra.MongoConfig{
		Username: "",
		Password: "",
		Host:     "localhost",
		Port:     27017,
		Database: "dev",
	}, &app.mainDB)
	redisCfg := infra.RedisConfig{
		Address:  "localhost:6379",
		Username: "",
		Password: "",
		Database: 0,
	}
	infraCtx.BindRedis(redisCfg, &app.mainCache)
	if err := infraCtx.Start(context.TODO()); err != nil {
		slog.Error("infrastructure startup failed", "error", err)
		os.Exit(1)
	}
	defer infraCtx.Shutdown(context.TODO())

	usersRepoCfg := usersrepo.MongoConfig{
		Database: app.mainDB.Database,
		Client:   app.mainDB.Client,
	}
	usersRepo := usersrepo.NewMongo(usersRepoCfg)

	usersServiceCfg := userssrv.Config{
		UsersRepo: usersRepo,
	}
	usersService := userssrv.New(usersServiceCfg)

	recipesRepoCfg := recipesrepo.MongoConfig{
		Database: app.mainDB.Database,
	}
	recipesRepo := recipesrepo.NewMongo(recipesRepoCfg)

	recipesServiceCfg := recipessrv.Config{
		RecipesRepo: recipesRepo,
		MaxLimit:    100,
	}

	recipesService := recipessrv.New(recipesServiceCfg)

	usersHandler := handlers.NewUsersHandler(usersService)
	recipesHandler := handlers.NewRecipesHandler(recipesService)

	authHandler := handlers.NewAuthHandler(usersService)

	router := gin.Default()

	// TODO: hardcoding connections and sceret
	secret := []byte("xnx6D7fCxR47XqHGrnkqIBDjHIoz1csJ")
	store, err := ginRedis.NewStore(10, "tcp", redisCfg.Address, redisCfg.Username, redisCfg.Password, secret)
	if err != nil {
		slog.Error("cache connection failed", "error", err)
		os.Exit(1)
	}
	// TODO: hardcoding session config
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	router.Use(sessions.Sessions("recipes", store))

	v1 := router.Group("/api/v1/")
	{
		v1.POST("/signup", usersHandler.SignUpHandler)
		v1.POST("/signin", authHandler.SignInHandler)
		v1.POST("/signout", authHandler.SignOutHandler)

		v1.GET("/recipes", recipesHandler.ListRecipesHandler)

		authed := v1.Group("/")
		authed.Use(authHandler.AuthMiddleware())
		{
			authed.POST("/recipes", recipesHandler.NewRecipeHandler)
			authed.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
			authed.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
		}
	}
	router.Run(":3000") // TODO: configurations
}
