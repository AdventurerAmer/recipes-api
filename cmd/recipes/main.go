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
	"time"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/AdventurerAmer/recipes-api/handlers"
	"github.com/AdventurerAmer/recipes-api/infra"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/recipessrv"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/userssrv"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/cache"
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

type AppConfig struct {
	AppName  string         `cffg:"appName"`
	Port     int            `cfg:"port"`
	Debug    bool           `cfg:"debug"`
	Database DatabaseConfig `cfg:"database"`
	Redis    RedisConfig    `cfg:"redis"`
}

type DatabaseConfig struct {
	URL      string `cfg:"url"`
	Username string `cfg:"username"`
	Password string `cfg:"password"`
	MaxConns int    `cfg:"maxConns"`
}

type RedisConfig struct {
	Host     string `cfg:"host"`
	Port     int    `cfg:"port"`
	Password string `cfg:"password"`
	DB       int    `cfg:"db"`
}

func main() {
	var cfg AppConfig
	if err := config.Load(&cfg); err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}
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

	ttl := 10 * time.Minute

	usersRepoCfg := usersrepo.MongoConfig{
		Database: app.mainDB.Database,
		Client:   app.mainDB.Client,
	}
	usersRepo := cache.NewRedisUsersRepository(usersrepo.NewMongo(usersRepoCfg), app.mainCache.Client, ttl)

	usersServiceCfg := userssrv.Config{
		UsersRepo: usersRepo,
	}
	usersService := userssrv.New(usersServiceCfg)

	recipesRepoCfg := recipesrepo.MongoConfig{
		Database: app.mainDB.Database,
	}
	recipesRepo := cache.NewRedisRecipesRepository(recipesrepo.NewMongo(recipesRepoCfg), app.mainCache.Client, ttl)

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
		v1.GET("/recipes/:id", recipesHandler.GetRecipeHandler)

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
