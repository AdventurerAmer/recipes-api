package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Exchange     string = "events"
	ExchangeType string = "topic"
	Queue        string = "events.work"
	RoutingKey   string = "events.#"
	ConsumerTag  string = "worker"
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

	// topic exchange is usually the best default for events
	if err := ch.ExchangeDeclare(Exchange, ExchangeType, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("'ch.ExchangeDeclare' failed: %w", err)
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

	routingKey := string(event.Name())

	if err := p.ch.PublishWithContext(ctx,
		Exchange,
		RoutingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    event.OccurredAt(),
			MessageId:    event.Key(),
			Type:         routingKey,
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("'ch.PublishWithContext' failed: %w", err)
	}

	return nil
}

type ampqSubscriber struct {
	conn    *amqp.Connection
	handler ports.EventHandler

	channel *amqp.Channel
	done    chan struct{}
	closed  bool
}

func NewAMPQSubscriber(conn *amqp.Connection, handler ports.EventHandler) ports.EventSubscriber {
	return &ampqSubscriber{
		conn:    conn,
		handler: handler,
		done:    make(chan struct{}),
	}
}

func (s *ampqSubscriber) Subscribe(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		default:
		}

		if err := s.connectAndConsume(ctx); err != nil {
			if s.isClosed() {
				return nil
			}
			log.Printf("consumer error: %v – reconnecting…", err)
			time.Sleep(2 * time.Second) // simple backoff; improve as needed
			continue
		}
	}
}

func (s *ampqSubscriber) connectAndConsume(ctx context.Context) error {
	if err := s.connect(); err != nil {
		return err
	}
	defer s.cleanup()

	if err := s.setupTopology(); err != nil {
		return err
	}

	_ = s.channel.Qos(1, 0, false)

	deliveries, err := s.channel.Consume(
		Queue,
		ConsumerTag,
		false, // manual ack → we control when the message is removed
		false, // exclusive = false → multiple consumers allowed
		false, false, nil,
	)
	if err != nil {
		return err
	}

	connClose := s.conn.NotifyClose(make(chan *amqp.Error, 1))
	chanClose := s.channel.NotifyClose(make(chan *amqp.Error, 1))

	log.Println("consumer ready")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		case err := <-connClose:
			return fmt.Errorf("connection closed: %v", err)
		case err := <-chanClose:
			return fmt.Errorf("channel closed: %v", err)
		case d, ok := <-deliveries:
			if !ok {
				return errors.New("deliveries closed")
			}
			s.process(ctx, d)
		}
	}
}

func (s *ampqSubscriber) process(ctx context.Context, d amqp.Delivery) {
	var event domain.Event
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("bad event, discarding: %v", err)
		_ = d.Nack(false, false) // drop poison message
		return
	}

	if err := s.handler(ctx, event); err != nil {
		log.Printf("handler error for %s: %v – requeue", event.Key(), err)
		_ = d.Nack(false, true) // requeue for another consumer (or later retry)
		return
	}

	_ = d.Ack(false) // message is permanently removed – only this consumer got it
}

func (s *ampqSubscriber) connect() error {
	ch, err := s.conn.Channel()
	if err != nil {
		_ = s.conn.Close()
		return err
	}
	s.channel = ch
	return nil
}

func (s *ampqSubscriber) setupTopology() error {
	if err := s.channel.ExchangeDeclare(
		Exchange, ExchangeType,
		true, false, false, false, nil,
	); err != nil {
		return err
	}

	_, err := s.channel.QueueDeclare(
		Queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, nil,
	)
	if err != nil {
		return err
	}

	return s.channel.QueueBind(
		Queue, RoutingKey, Exchange, false, nil,
	)
}

func (s *ampqSubscriber) cleanup() {
	if s.channel != nil {
		_ = s.channel.Close()
	}
}

func (s *ampqSubscriber) Close() {
	if !s.closed {
		s.closed = true
		close(s.done)
	}
	s.cleanup()
}

func (c *ampqSubscriber) isClosed() bool { return c.closed }
