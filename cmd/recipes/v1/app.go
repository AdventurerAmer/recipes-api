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
	"github.com/AdventurerAmer/recipes-api/internal/adapters/password"
	"github.com/AdventurerAmer/recipes-api/internal/adapters/textsearch"
	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/authsrv"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/recipessrv"
	"github.com/AdventurerAmer/recipes-api/internal/core/services/userssrv"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/recipesrepo"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/usersrepo"
	"github.com/AdventurerAmer/recipes-api/logging"
	"github.com/AdventurerAmer/recipes-api/mongoutils"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gin-contrib/sessions"
	ginRedis "github.com/gin-contrib/sessions/redis"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New()
		logger.Error("failed to load config", "error", err)
		return 1
	}

	serviceCfg := cfg.Services.Recipes

	logger := cfg.NewLogger().With(slog.String("service", serviceCfg.Name))

	var (
		mainDataBase      infra.MongoContext
		mainCache         infra.RedisContext
		mainMessageBroker infra.RabbitMqContext
		mainObjectStorage infra.MinioContext
		mainTextSearch    infra.ElasticSearchContext
	)

	inf := infra.New()
	inf.BindMongo(&cfg.Infra.MainDatabase, &mainDataBase)
	inf.BindRedis(&cfg.Infra.MainCache, &mainCache)
	inf.BindRabbitMQ(&cfg.Infra.MainMessageBroker, &mainMessageBroker)
	inf.BindMinio(&cfg.Infra.MainObjectStorage, &mainObjectStorage)
	inf.BindElasticSearch(&cfg.Infra.MainTextSearch, &mainTextSearch)
	if err := inf.Start(context.Background()); err != nil {
		logger.Error("failed to connect to infrastructure", "error", err)
		return 1
	}
	defer inf.Shutdown(context.Background())

	transactor := mongoutils.NewTransactor(mainDataBase.Client)

	redisCache := cache.NewRedis(mainCache.Client)

	textSearch, err := textsearch.NewElasticSearch(mainTextSearch.Client)
	if err != nil {
		logger.Error("failed to create elastic search adaptor", "error", err)
		return 1
	}

	// Repos
	usersRepoCfg := usersrepo.MongoConfig{
		Database:   mainDataBase.Database,
		Cache:      redisCache,
		Transactor: transactor,
	}
	usersRepo := usersrepo.NewMongo(usersRepoCfg)

	recipesRepoCfg := recipesrepo.MongoConfig{
		Database:   mainDataBase.Database,
		TextSearch: textSearch,
		Cache:      redisCache,
		Transactor: transactor,
	}
	recipesRepo := recipesrepo.NewMongo(recipesRepoCfg)

	// Services
	argon2PasswordMgr := password.NewArgon2()

	authServiceCfg := &authsrv.Config{
		PasswordVerifier: password.NewArgon2(),
		UsersRepo:        usersRepo,
	}

	authService := authsrv.New(authServiceCfg)

	usersServiceCfg := userssrv.Config{
		PasswordHasher: argon2PasswordMgr,
		UsersRepo:      usersRepo,
		Transactor:     transactor,
		EventPublisher: mainMessageBroker.Client,
	}
	usersService := userssrv.New(usersServiceCfg)

	recipesServiceCfg := recipessrv.Config{
		RecipesRepo: recipesRepo,
		MaxLimit:    100, // TODO: hardcoding
	}
	recipesService := recipessrv.New(recipesServiceCfg)

	// Handlers
	authHandler := handlers.NewAuth(authService)
	usersHandler := handlers.NewUsers(usersService)
	recipesHandler := handlers.NewRecipes(recipesService)

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

	v1.POST("/events/user-created", func(c *gin.Context) {
		event := domain.NewUserCreated(uuid.NewString())
		if err := mainMessageBroker.Client.Publish(c, event); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		slog.Info("published an event", "name", event.Name(), "userId", event.UserId)
	})

	{
		v1.POST("/users", usersHandler.Register)
		v1.POST("/sessions", authHandler.Login)
		v1.POST("/sessions/current", authHandler.Logout)

		v1.GET("/recipes", recipesHandler.List)
		v1.GET("/recipes/search", recipesHandler.Search)
		v1.GET("/recipes/:id", recipesHandler.Get)

		authed := v1.Group("/")
		authed.Use(authHandler.AuthMiddleware())
		{
			authed.POST("/recipes", recipesHandler.Create)
			authed.PUT("/recipes/:id", recipesHandler.Update)
			authed.DELETE("/recipes/:id", recipesHandler.Delete)
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
