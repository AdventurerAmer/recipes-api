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
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/AdventurerAmer/recipes-api/handlers"
	"github.com/AdventurerAmer/recipes-api/infra"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/recipessrv"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/userssrv"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/cache"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/recipesrepo"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/usersrepo"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"
	ginRedis "github.com/gin-contrib/sessions/redis"
)

type InfraCfg struct {
	MainDB        infra.MongoConfig `cfg:"mainDatabase"`
	MainCache     infra.RedisConfig `cfg:"mainCache"`
	SessionsCache infra.RedisConfig `cfg:"sessionsCache"`
}

type SessionsCfg struct {
	Secret       string        `cfg:"secret"`
	MaxIdelConns int           `cfg:"maxIdelConns"`
	Name         string        `cfg:"name"`
	MaxAge       time.Duration `cfg:"maxAge"`
}

type ServerCfg struct {
	Port     int         `cfg:"port"`
	Sessions SessionsCfg `cfg:"sessions"`
}

type Constants struct {
	DefaultTimeout  time.Duration `cfg:"defaultTimeout"`
	UsersCacheTTL   time.Duration `cfg:"usersCacheTTL"`
	RecipesCacheTTL time.Duration `cfg:"recipesCacheTTL"`
	RecipesMaxLimit int           `cfg:"recipesMaxLimit"`
}

type Config struct {
	Infra     InfraCfg  `cfg:"infra"`
	Server    ServerCfg `cfg:"server"`
	Constants Constants `cfg:"constants"`
}

type App struct {
	mainDB        infra.MongoContext
	mainCache     infra.RedisContext
	sessionsCache infra.RedisContext
}

func main() {
	var cfg Config
	if err := config.Load(&cfg, config.WithPrefix("RECIPES")); err != nil {
		slog.Error("load configuation failed", "error", err)
		os.Exit(1)
	}

	slog.Info("configuation loaded successfully", "data", cfg)

	app := &App{}
	infraCtx := infra.New()
	infraCtx.BindMongo(cfg.Infra.MainDB, &app.mainDB)
	infraCtx.BindRedis(cfg.Infra.MainCache, &app.mainCache)
	if err := infraCtx.Start(context.TODO()); err != nil {
		slog.Error("infrastructure startup failed", "error", err)
		os.Exit(1)
	}
	defer infraCtx.Shutdown(context.TODO())

	usersRepoCfg := usersrepo.MongoConfig{
		Database: app.mainDB.Database,
		Client:   app.mainDB.Client,
	}
	usersRepo := cache.NewRedisUsersRepository(usersrepo.NewMongo(usersRepoCfg), app.mainCache.Client, cfg.Constants.UsersCacheTTL)

	usersServiceCfg := userssrv.Config{
		UsersRepo: usersRepo,
	}
	usersService := userssrv.New(usersServiceCfg)

	recipesRepoCfg := recipesrepo.MongoConfig{
		Database: app.mainDB.Database,
	}
	recipesRepo := cache.NewRedisRecipesRepository(recipesrepo.NewMongo(recipesRepoCfg), app.mainCache.Client, cfg.Constants.RecipesCacheTTL)

	recipesServiceCfg := recipessrv.Config{
		RecipesRepo: recipesRepo,
		MaxLimit:    cfg.Constants.RecipesMaxLimit,
	}

	recipesService := recipessrv.New(recipesServiceCfg)

	usersHandler := handlers.NewUsersHandler(usersService)
	recipesHandler := handlers.NewRecipesHandler(recipesService)

	authHandler := handlers.NewAuthHandler(usersService)

	secret := []byte(cfg.Server.Sessions.Secret)
	sessionsStore, err := ginRedis.NewStore(cfg.Server.Sessions.MaxIdelConns, "tcp", cfg.Infra.SessionsCache.Address, cfg.Infra.SessionsCache.Username, cfg.Infra.SessionsCache.Password, secret)
	if err != nil {
		slog.Error("cache connection failed", "error", err)
		os.Exit(1)
	}

	// TODO: using dev config for new
	sessionsStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int(cfg.Server.Sessions.MaxAge.Seconds()),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	router := gin.Default()
	router.Use(sessions.Sessions(cfg.Server.Sessions.Name, sessionsStore))

	v1 := router.Group("/api/v1/")
	v1.Use(timeout.New(timeout.WithTimeout(cfg.Constants.DefaultTimeout)))
	{
		v1.POST("/signup", usersHandler.SignUpHandler)
		v1.POST("/signin", authHandler.SignInHandler)
		v1.POST("/signout", authHandler.SignOutHandler)

		v1.GET("/recipes", recipesHandler.ListRecipesHandler)
		v1.GET("/recipes/:id", recipesHandler.GetRecipeHandler)

		authed := v1.Group("/")
		authed.Use(authHandler.AuthMiddleware())
		{
			authed.POST("/recipes", recipesHandler.NewRecipeHandler)
			authed.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
			authed.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
		}
	}
	router.Run(fmt.Sprintf(":%d", cfg.Server.Port))
}
