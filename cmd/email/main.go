package email

import (
	"context"
	"log/slog"
	"os"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/AdventurerAmer/recipes-api/infra"
	"github.com/AdventurerAmer/recipes-api/logging"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New()
		logger.Error("failed to load config", "error", err)
		return 1
	}

	logger := cfg.NewLogger().With(slog.String("worker", "email"))
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
	return 0
}

func main() {
	os.Exit(Run())
}
