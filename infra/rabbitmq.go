package infra

import (
	"context"
	"fmt"
	"net"

	"github.com/AdventurerAmer/recipes-api/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMq struct {
	*config.RabbitMq
}

func (cfg *RabbitMq) Connect(ctx context.Context) (Disconnecter, error) {
	config := amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{}
			return dialer.Dial(network, addr)
		},
	}
	conn, err := amqp.DialConfig(cfg.Addr(), config)
	if err != nil {
		return nil, fmt.Errorf("'amqp.DialConfig' failed: %w", err)
	}
	return &RabbitMqContext{
		Connection: conn,
	}, nil
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
