package infra

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMq struct {
	*config.RabbitMq
}

func (cfg *RabbitMq) Connect(ctx context.Context) (Disconnecter, error) {
	connStr := fmt.Sprintf("amqp://%s:%s@%s", cfg.Username, cfg.Password, cfg.Addr())
	type result struct {
		conn *amqp.Connection
		err  error
	}
	ch := make(chan result)
	go func() {
		conn, err := amqp.Dial(connStr)
		if err != nil {
			err = fmt.Errorf("'amqp.Dial' failed: %w", err)
		}
		ch <- result{
			conn: conn,
			err:  err,
		}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return nil, res.err
		}
		return &RabbitMqContext{
			Connection: res.conn,
		}, nil
	}
}

type RabbitMqContext struct {
	Connection *amqp.Connection
}

func (r *RabbitMqContext) Disconnect(ctx context.Context) error {
	if err := r.Connection.Close(); err != nil {
		return fmt.Errorf("'Connection.Close' failed: %w", err)
	}
	r.Connection = nil
	return nil
}
