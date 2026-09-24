package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/AdventurerAmer/recipes-api/infra"
	"github.com/AdventurerAmer/recipes-api/internal/adapters/broker"
	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/logging"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New()
		logger.Error("failed to load config", "error", err)
		return 1
	}

	logger := cfg.NewLogger().With(slog.String("worker", "another"))
	var (
		mainDataBase      infra.MongoContext
		mainCache         infra.RedisContext
		mainMessageBroker infra.RabbitMqContext
	)

	inf := infra.New()
	inf.BindMongo(&cfg.Infra.MainDatabase, &mainDataBase)
	inf.BindRedis(&cfg.Infra.MainCache, &mainCache)
	inf.BindRabbitMQ(&cfg.Infra.MainMessageBroker, &mainMessageBroker)
	if err := inf.Start(context.Background()); err != nil {
		logger.Error("failed to connect to infrastructure", "error", err)
		return 1
	}
	defer inf.Shutdown(context.Background())

	subCfg := &broker.AMQPSubscriberConfig{
		Name: "another",
		Conn: mainMessageBroker.Connection,
		Handler: func(ctx context.Context, event domain.Event) error {
			return nil
		},
	}
	sub, err := broker.NewAMPQSubscriber(subCfg, domain.EventNameUserCreated, domain.EventNameUserPasswordReset, domain.EventNameUserVerification)
	if err != nil {
		logger.Error("failed to create ampq adptor", "error", err)
		return 1
	}

	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()

	if err := sub.Subscribe(sigCtx); err != nil {
		logger.Error("failed to create subscribe to events", "error", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(Run())
}
