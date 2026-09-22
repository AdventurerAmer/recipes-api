package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ampqPublisher struct {
	ch       *amqp.Channel
	exchange string
}

func NewAMPQPublisher(conn *amqp.Connection, exchange string) (ports.EventPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("'conn.Channel' failed: %w", err)
	}

	// topic exchange is usually the best default for events
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("'ch.ExchangeDeclare' failed: %w", err)
	}

	return &ampqPublisher{
		ch:       ch,
		exchange: exchange,
	}, nil
}

func (p *ampqPublisher) Publish(ctx context.Context, event domain.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}

	routingKey := event.Name()

	if err := p.ch.PublishWithContext(ctx,
		p.exchange,
		string(routingKey),
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    event.OccurredAt(),
			MessageId:    event.Id(),
			Type:         string(event.Name()),
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("'ch.PublishWithContext' failed: %w", err)
	}

	return nil
}
