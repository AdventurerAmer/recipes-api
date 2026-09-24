package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ampqPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewAMPQPublisher(conn *amqp.Connection) (ports.EventPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("'conn.Channel' failed: %w", err)
	}

	if err := ensureTopology(ch, domain.EventNames); err != nil {
		return nil, fmt.Errorf("'ensureTopology' failed: %w", err)
	}

	return &ampqPublisher{
		conn: conn,
		ch:   ch,
	}, nil
}

func (p *ampqPublisher) Publish(ctx context.Context, event domain.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}

	eventName := event.Name().String()
	exhange := eventName

	if err := p.ch.PublishWithContext(ctx,
		exhange,
		"",
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    event.OccurredAt(),
			MessageId:    event.Key(),
			Type:         eventName,
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("'ch.PublishWithContext' failed: %w", err)
	}

	return nil
}

type AMQPSubscriberConfig struct {
	Name     string
	Conn     *amqp.Connection
	Registry ports.EventRegistry
}

type ampqSubscriber struct {
	name     string
	conn     *amqp.Connection
	channel  *amqp.Channel
	registry ports.EventRegistry
}

func NewAMPQSubscriber(cfg *AMQPSubscriberConfig) (ports.EventSubscriber, error) {
	ch, err := cfg.Conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("'Conn.Channel' failed: %w", err)
	}

	eventNames := slices.Collect(maps.Keys(cfg.Registry))

	if err := ensureTopology(ch, eventNames); err != nil {
		return nil, fmt.Errorf("'ensureTopology' failed: %w", err)
	}

	durable := true
	queue, err := ch.QueueDeclare(
		cfg.Name, // name
		durable,  // durability
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("'ch.QueueDeclare' failed: %w", err)
	}

	for _, eventName := range eventNames {
		if err := ch.QueueBind(
			queue.Name,
			"",
			eventName.String(),
			false,
			nil); err != nil {
			return nil, fmt.Errorf("'ch.QueueBind' failed: %w", err)
		}
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("'ch.Qos' failed: %w", err)
	}

	return &ampqSubscriber{
		name:     cfg.Name,
		registry: cfg.Registry,
		conn:     cfg.Conn,
		channel:  ch,
	}, nil
}

func (s *ampqSubscriber) Start(ctx context.Context) error {
	msgs, err := s.channel.Consume(
		s.name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("'channel.Consume' failed: %w", err)
	}
	for d := range msgs {
		// key := d.MessageId
		eventName := domain.EventName(d.Type)
		// occurredAt := d.Timestamp
		handler := s.registry[eventName]
		event, err := handler.Unmarshal(d.Body)
		if err != nil {
			// TODO: log here
			continue
		}
		if err := handler.Handle(ctx, event); err != nil {
			// handle error here...
			_ = d.Nack(false, true)
		}
		_ = d.Ack(false)
	}
	return nil
}

func ensureTopology(ch *amqp.Channel, eventNames []domain.EventName) error {
	for _, eventName := range eventNames {
		exchange := eventName.String()
		durable := true
		if err := ch.ExchangeDeclare(exchange, "fanout", durable, false, false, false, nil); err != nil {
			return fmt.Errorf("'ch.ExchangeDeclare' failed: %w", err)
		}
	}
	return nil
}
