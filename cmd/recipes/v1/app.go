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
package v1

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdventurerAmer/recipes-api/cmd/recipes/v1/handlers"
	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/AdventurerAmer/recipes-api/infra"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/recipessrv"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/userssrv"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/cache"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/recipesrepo"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/usersrepo"
	"github.com/AdventurerAmer/recipes-api/internal/text_searches/elasticsearch"
	"github.com/AdventurerAmer/recipes-api/logging"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"
	ginRedis "github.com/gin-contrib/sessions/redis"
)

type App struct {
	mainDB            infra.MongoContext
	mainCache         infra.RedisContext
	sessionsCache     infra.RedisContext
	mainObjectStorage infra.MinioContext
	mainTextSearch    infra.ElasticSearchContext
}

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New(nil)
		logger.Error("failed to load config", "error", err)
		return 1
	}

	serviceCfg := cfg.Services.Recipes

	format := "json"
	if cfg.Env == config.EnvLocal {
		format = "text"
	}
	loggingCfg := logging.Config{
		IsLocalEnv: cfg.Env == config.EnvLocal,
		Level:      logging.ParseLevel(cfg.Observability.Logging.Level),
		AddSource:  cfg.Env != config.EnvProduction,
		Format:     format,
	}
	logger := logging.New(&loggingCfg).With(slog.String("service", serviceCfg.Name))

	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()

	app := &App{}
	infraCtx := infra.New()
	infraCtx.BindMongo(&cfg.Infra.MainDatabase, &app.mainDB)
	infraCtx.BindRedis(&cfg.Infra.MainCache, &app.mainCache)
	infraCtx.BindMinio(&cfg.Infra.MainObjectStorage, &app.mainObjectStorage)
	infraCtx.BindElasticSearch(&cfg.Infra.MainTextSearch, &app.mainTextSearch)
	if err := infraCtx.Start(sigCtx); err != nil {
		logger.Error("failed to connect to infrastructure", "error", err)
		return 1
	}
	defer infraCtx.Shutdown(context.Background())

	usersRepoCfg := usersrepo.MongoConfig{
		Database: app.mainDB.Database,
		Client:   app.mainDB.Client,
	}
	usersRepo := usersrepo.NewMongo(usersRepoCfg)
	usersRepo = cache.NewRedisUsersRepository(usersRepo,
		app.mainCache.Client, 10*time.Minute) // TODO:hardcoding

	usersServiceCfg := userssrv.Config{
		UsersRepo: usersRepo,
	}
	usersService := userssrv.New(usersServiceCfg)

	textSearch, err := elasticsearch.New(app.mainTextSearch.Client)
	if err != nil {
		logger.Error("failed to create elastic search port", "error", err)
		return 1
	}

	recipesRepoCfg := recipesrepo.MongoConfig{
		Database:   app.mainDB.Database,
		Client:     app.mainDB.Client,
		TextSearch: textSearch,
	}
	recipesRepo := recipesrepo.NewMongo(recipesRepoCfg)
	recipesRepo = cache.NewRedisRecipesRepository(recipesRepo, app.mainCache.Client, 10*time.Minute) // TODO: hardcoding

	recipesServiceCfg := recipessrv.Config{
		RecipesRepo: recipesRepo,
		MaxLimit:    100, // TODO: hardcoding
	}
	recipesService := recipessrv.New(recipesServiceCfg)

	usersHandler := handlers.NewUsersHandler(usersService)
	recipesHandler := handlers.NewRecipesHandler(recipesService)
	authHandler := handlers.NewAuthHandler(usersService)

	secret := []byte(cfg.Auth.Secret)
	sessionsStore, err := ginRedis.NewStore(cfg.Auth.MaxIdelConns, "tcp", cfg.Infra.SessionsCache.Addr(), cfg.Infra.SessionsCache.Username, cfg.Infra.SessionsCache.Password, secret)
	if err != nil {
		logger.Error("failed to create redis sessions store", "error", err)
		return 1
	}

	// TODO: using dev config for new
	sessionsStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int(cfg.Auth.MaxAge.Seconds()),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	router := gin.Default()
	router.Use(sessions.Sessions(cfg.Auth.Name, sessionsStore))

	v1 := router.Group("/api/v1/")
	v1.Use(timeout.New(timeout.WithTimeout(cfg.Services.Recipes.DefaultTimeout)))
	{
		v1.POST("/signup", usersHandler.SignUpHandler)
		v1.POST("/signin", authHandler.SignInHandler)
		v1.POST("/signout", authHandler.SignOutHandler)

		v1.GET("/recipes", recipesHandler.ListRecipesHandler)
		v1.GET("/recipes/search", recipesHandler.SearchRecipesHandler)
		v1.GET("/recipes/:id", recipesHandler.GetRecipeHandler)

		authed := v1.Group("/")
		authed.Use(authHandler.AuthMiddleware())
		{
			authed.POST("/recipes", recipesHandler.NewRecipeHandler)
			authed.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
			authed.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
		}
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Services.Recipes.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("'srv.ListenAndServe' failed", "error", err)
		}
	}()
	<-sigCtx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("'srv.Shutdown' failed", "error", err)
		return 1
	}

	logger.Info("Gracefully shutdown was successful")

	return 0
}
