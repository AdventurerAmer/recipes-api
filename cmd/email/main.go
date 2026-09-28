package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/AdventurerAmer/recipes-api/infrastructure"
	"github.com/AdventurerAmer/recipes-api/internal/adapters/broker"
	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/AdventurerAmer/recipes-api/logging"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) OnUserCreated(ctx context.Context, e domain.UserCreatedEvent) error {
	slog.Info("OnUserCreated")
	return nil
}

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New()
		logger.Error("failed to load config", "error", err)
		return 1
	}

	logger := cfg.NewLogger().With(slog.String("worker", "email"))

	var (
		mainDataBase      infrastructure.MongoContext
		mainCache         infrastructure.RedisContext
		mainMessageBroker infrastructure.RabbitMqContext
	)

	infra, err := infrastructure.New(logger)
	if err != nil {
		logger.Error("failed to connect to infrastructure", "error", err)
		return 1
	}

	infra.BindMongo(&cfg.Infra.MainDatabase, &mainDataBase)
	infra.BindRedis(&cfg.Infra.MainCache, &mainCache)
	infra.BindRabbitMQ(&cfg.Infra.MainMessageBroker, &mainMessageBroker)
	if err := infra.Start(context.Background()); err != nil {
		logger.Error("failed to connect to infrastructure", "error", err)
		return 1
	}
	defer infra.Shutdown(context.Background())

	h := NewHandler()
	dispatcher := ports.NewEventDispatcher()
	ports.RegisterEvent(dispatcher, domain.EventNameUserCreated, h.OnUserCreated)
	// TODO: move to go 1.27 for this to be
	// dispatcher.Register(domain.EventNameUserCreated, h.OnUserCreated)

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

func main() {
	os.Exit(Run())
}
