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
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdventurerAmer/recipes-api/cmd/recipes/v1/handlers"
	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/AdventurerAmer/recipes-api/infra"
	"github.com/AdventurerAmer/recipes-api/internal/adapters/cache"
	"github.com/AdventurerAmer/recipes-api/internal/adapters/textsearch"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/recipessrv"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/userssrv"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/recipesrepo"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/usersrepo"
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
		logger := logging.New()
		logger.Error("failed to load config", "error", err)
		return 1
	}

	serviceCfg := cfg.Services.Recipes

	logger := cfg.NewLogger().With(slog.String("service", serviceCfg.Name))

	app := &App{}
	infraCtx := infra.New()
	infraCtx.BindMongo(&cfg.Infra.MainDatabase, &app.mainDB)
	infraCtx.BindRedis(&cfg.Infra.MainCache, &app.mainCache)
	infraCtx.BindMinio(&cfg.Infra.MainObjectStorage, &app.mainObjectStorage)
	infraCtx.BindElasticSearch(&cfg.Infra.MainTextSearch, &app.mainTextSearch)
	if err := infraCtx.Start(context.Background()); err != nil {
		logger.Error("failed to connect to infrastructure", "error", err)
		return 1
	}
	defer infraCtx.Shutdown(context.Background())

	redisCache := cache.NewRedis(app.mainCache.Client)

	textSearch, err := textsearch.NewElasticSearch(app.mainTextSearch.Client)
	if err != nil {
		logger.Error("failed to create elastic search port", "error", err)
		return 1
	}

	usersRepoCfg := usersrepo.MongoConfig{
		Database: app.mainDB.Database,
		Client:   app.mainDB.Client,
		Cache:    redisCache,
	}
	usersRepo := usersrepo.NewMongo(usersRepoCfg)

	usersServiceCfg := userssrv.Config{
		UsersRepo: usersRepo,
	}
	usersService := userssrv.New(usersServiceCfg)

	recipesRepoCfg := recipesrepo.MongoConfig{
		Database:   app.mainDB.Database,
		Client:     app.mainDB.Client,
		TextSearch: textSearch,
		Cache:      redisCache,
	}
	recipesRepo := recipesrepo.NewMongo(recipesRepoCfg)

	recipesServiceCfg := recipessrv.Config{
		RecipesRepo: recipesRepo,
		MaxLimit:    100, // TODO: hardcoding
	}
	recipesService := recipessrv.New(recipesServiceCfg)

	usersHandler := handlers.NewUsersHandler(usersService)
	recipesHandler := handlers.NewRecipesHandler(recipesService)
	authHandler := handlers.NewAuthHandler(usersService)

	sessionsStore, err := ginRedis.NewStore(cfg.Auth.MaxIdelConns, "tcp", cfg.Infra.SessionsCache.Addr(), cfg.Infra.SessionsCache.Username, cfg.Infra.SessionsCache.Password, []byte(cfg.Auth.Secret))
	if err != nil {
		logger.Error("failed to create redis sessions store", "error", err)
		return 1
	}

	sessionStoreOpts := sessions.Options{
		Path:     "/",
		MaxAge:   int(cfg.Auth.MaxAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	sessionsStore.Options(sessionStoreOpts)

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
		Addr:              serviceCfg.Addr(),
		MaxHeaderBytes:    serviceCfg.MaxHeaderBytes,
		ReadHeaderTimeout: serviceCfg.ReadHeaderTimeout,
		ReadTimeout:       serviceCfg.ReadTimeout,
		WriteTimeout:      serviceCfg.WriteTimeout,
		IdleTimeout:       serviceCfg.IdleTimeout,
		Handler:           router,
	}

	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()

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
