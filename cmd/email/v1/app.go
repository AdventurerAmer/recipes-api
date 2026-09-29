package v1

import (
	"context"
	"html/template"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/AdventurerAmer/recipes-api/infrastructure"
	"github.com/AdventurerAmer/recipes-api/internal/adapters/broker"
	"github.com/AdventurerAmer/recipes-api/internal/adapters/cache"
	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/AdventurerAmer/recipes-api/internal/repositories/usersrepo"
	"github.com/AdventurerAmer/recipes-api/logging"
	"github.com/AdventurerAmer/recipes-api/mailer"
	"github.com/AdventurerAmer/recipes-api/mongoutils"
)

func Run(templates *template.Template) int {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New()
		logger.Error("failed to load config", "error", err)
		return 1
	}

	logger := cfg.NewLogger().With(slog.String("worker", "email"))

	infra, err := infrastructure.New(logger)
	if err != nil {
		logger.Error("failed to connect to infrastructure", "error", err)
		return 1
	}

	mainDataBase := infra.Mongo(&cfg.Infra.MainDatabase)
	_ = mainDataBase

	mainCache := infra.Redis(&cfg.Infra.MainCache)
	_ = mainCache

	mainMessageBroker := infra.RabbitMQ(&cfg.Infra.MainMessageBroker)
	if err := infra.Start(context.Background()); err != nil {
		logger.Error("failed to connect to infrastructure", "error", err)
		return 1
	}
	defer infra.Shutdown(context.Background())

	transactor := mongoutils.NewTransactor(mainDataBase.Client)

	redisCache := cache.NewRedis(mainCache.Client)

	usersRepoCfg := usersrepo.MongoConfig{
		Database:   mainDataBase.Database,
		Cache:      redisCache,
		Transactor: transactor,
	}
	usersRepo := usersrepo.NewMongo(usersRepoCfg)

	mailer := mailer.New(&cfg.Mailer)
	h := newEventHandler(usersRepo, templates, mailer)
	// TODO: move to go 1.27 for this to be
	// dispatcher.Register(domain.EventNameUserCreated, h.OnUserCreated)
	dispatcher := ports.NewEventDispatcher()
	ports.RegisterEvent(dispatcher, domain.EventNameUserCreated, h.OnUserCreated)
	ports.RegisterEvent(dispatcher, domain.EventNameUserVerification, h.OnUserVerification)
	ports.RegisterEvent(dispatcher, domain.EventNameUserPasswordReset, h.OnUserPasswordReset)

	consumerCfg := broker.AMQPConsumerConfig{
		Name:        "email",
		Dispatcher:  dispatcher,
		WorkerCount: 128,
		Timeout:     5 * time.Second,
		AckTimeout:  2 * time.Second,
	}
	consumer := broker.NewAMQPConsumer(consumerCfg, mainMessageBroker.Client)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	consumer.Start()

	<-ctx.Done()

	consumer.Stop()

	return 0
}
