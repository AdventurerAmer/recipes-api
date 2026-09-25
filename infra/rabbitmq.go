package infra

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/AdventurerAmer/recipes-api/internal/adapters/broker"
)

type RabbitMq struct {
	*config.RabbitMq
}

func (cfg *RabbitMq) Connect(ctx context.Context) (Disconnecter, error) {
	connStr := fmt.Sprintf("amqp://%s:%s@%s", cfg.Username, cfg.Password, cfg.Addr())
	type result struct {
		ctx *RabbitMqContext
		err error
	}
	ch := make(chan result)
	go func() {
		client, err := broker.NewAMQPClient(connStr)
		if err != nil {
			ch <- result{err: fmt.Errorf("'broker.NewAMQPClient' failed: %w", err)}
			return
		}
		ch <- result{ctx: &RabbitMqContext{Client: client}, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return nil, res.err
		}
		return res.ctx, nil
	}
}

type RabbitMqContext struct {
	Client *broker.AMQPClient
}

func (c *RabbitMqContext) Disconnect(ctx context.Context) error {
	if err := c.Client.Close(); err != nil {
		return fmt.Errorf("'Channel.Close' failed: %w", err)
	}
	return nil
}
