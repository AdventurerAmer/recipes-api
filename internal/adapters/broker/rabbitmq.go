package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/AdventurerAmer/recipes-api/logging"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPClient struct {
	connStr string

	mu            sync.RWMutex
	conn          *amqp.Connection
	publishMu     sync.Mutex
	publishCh     *amqp.Channel
	consumeCh     *amqp.Channel
	done          chan struct{}
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	isConnected   bool

	logger *logging.Logger
}

func NewAMQPClient(connStr string, logger *logging.Logger) (*AMQPClient, error) {
	c := &AMQPClient{
		connStr: connStr,
		done:    make(chan struct{}),
	}
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("'client.connect' failed: %w", err)
	}

	go c.handleReconnect()

	return c, nil
}

func (c *AMQPClient) Publish(ctx context.Context, event domain.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}

	c.publishMu.Lock()
	defer c.publishMu.Unlock()

	for {
		c.mu.RLock()
		connected := c.isConnected
		ch := c.publishCh
		confirms := c.notifyConfirm
		c.mu.RUnlock()

		if !connected || ch == nil || confirms == nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}

		eventName := event.Name().String()
		exchange := eventName

		// Attempt to publish
		if err := ch.PublishWithContext(
			ctx,
			exchange,
			"",
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Type:         eventName,
				MessageId:    event.Id(),
				Timestamp:    event.OccurredAt(),
				Body:         body,
			},
		); err != nil {
			c.logger.Error("Publish failed", "error", err)
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case confirm, ok := <-confirms:
			if !ok {
				continue
			}
			if confirm.Ack {
				return nil
			}
			c.logger.Error("Publish Nacked")
			continue
		}
	}
}

func (c *AMQPClient) Close() error {
	close(c.done)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.publishCh != nil {
		if err := c.publishCh.Close(); err != nil {
			c.logger.Error("Close publish channel failed", "error", err)
		}
	}

	if c.consumeCh != nil {
		if err := c.consumeCh.Close(); err != nil {
			c.logger.Error("Close consume channel failed", "error", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("'conn.Close' failed: %w", err)
		}
	}

	c.isConnected = false

	c.logger.Info("Client cLosed")

	return nil
}

func (c *AMQPClient) connect() error {
	cfg := amqp.Config{
		Heartbeat: 10 * time.Second,
	}
	conn, err := amqp.DialConfig(c.connStr, cfg)
	if err != nil {
		return fmt.Errorf("'amqp.DialConfig' failed: %w", err)
	}

	publishCh, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("'conn.Channel' failed: %w", err)
	}

	consumeCh, err := conn.Channel()
	if err != nil {
		_ = publishCh.Close()
		_ = conn.Close()
		return fmt.Errorf("'conn.Channel' failed: %w", err)
	}

	if err := publishCh.Confirm(false); err != nil {
		_ = consumeCh.Close()
		_ = publishCh.Close()
		_ = conn.Close()
		return fmt.Errorf("'publishCh.Confirm' failed: %w", err)
	}

	for _, eventName := range domain.EventNames {
		exchange := eventName.String()
		durable := true
		if err := publishCh.ExchangeDeclare(
			exchange,
			"fanout",
			durable,
			false,
			false,
			false,
			nil); err != nil {
			_ = consumeCh.Close()
			_ = publishCh.Close()
			_ = conn.Close()
			return fmt.Errorf("'publishCh.ExchangeDeclare' failed: %w", err)
		}
	}

	notifyClose := make(chan *amqp.Error, 1)
	notifyConfirm := make(chan amqp.Confirmation, 1)
	conn.NotifyClose(notifyClose)
	publishCh.NotifyPublish(notifyConfirm)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn = conn
	c.publishCh = publishCh
	c.consumeCh = consumeCh
	c.notifyClose = notifyClose
	c.notifyConfirm = notifyConfirm
	c.isConnected = true

	return nil
}

func (c *AMQPClient) handleReconnect() {
	for {
		select {
		case <-c.done:
			return
		case err := <-c.notifyClose:
			if err != nil {
				c.logger.Error("Connection was closed", "error", err)
			}

			c.mu.Lock()
			c.isConnected = false
			c.mu.Unlock()

			c.reconnectWithBackoff()
		}
	}
}

func (c *AMQPClient) reconnectWithBackoff() {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-c.done:
			return
		default:
		}

		time.Sleep(backoff)

		if err := c.connect(); err != nil {
			c.logger.Error("Failed to reconnect", "error", err)

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		c.logger.Info("Reconnected successfully")
		return
	}
}

type AMQPConsumerConfig struct {
	Name        string
	Dispatcher  *ports.EventDispatcher
	Timeout     time.Duration
	AckTimeout  time.Duration
	WorkerCount int
}

type AMQPConsumer struct {
	AMQPConsumerConfig
	client     *AMQPClient
	deliveryCh chan amqp.Delivery
	done       chan struct{}
	wg         sync.WaitGroup
}

func NewAMQPConsumer(cfg AMQPConsumerConfig, client *AMQPClient) *AMQPConsumer {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.AckTimeout == 0 {
		cfg.AckTimeout = 2 * time.Second
	}
	return &AMQPConsumer{
		AMQPConsumerConfig: cfg,
		client:             client,
		deliveryCh:         make(chan amqp.Delivery, 1024),
		done:               make(chan struct{}),
	}
}

func (c *AMQPConsumer) Start() {
	c.wg.Go(c.consumeLoop)
	for range c.WorkerCount {
		c.wg.Go(func() {
			for msg := range c.deliveryCh {
				c.processMessage(msg)
			}
		})
	}
}

func (c *AMQPConsumer) Stop() {
	close(c.done)
	close(c.deliveryCh)
	c.wg.Wait()
}

func (c *AMQPConsumer) consumeLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
		}

		c.client.mu.RLock()
		connected := c.client.isConnected
		ch := c.client.consumeCh
		c.client.mu.RUnlock()

		if !connected || ch == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		eventNames := c.Dispatcher.Events()

		durable := true
		queue, err := ch.QueueDeclare(
			c.Name,  // name
			durable, // durability
			false,   // delete when unused
			false,   // exclusive
			false,   // no-wait
			nil,     // arguments
		)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		for _, eventName := range eventNames {
			exchange := eventName
			if err := ch.QueueBind(
				queue.Name,
				"",
				exchange,
				false,
				nil); err != nil {
				time.Sleep(time.Second)
				continue
			}
		}

		if err := ch.Qos(1, 0, false); err != nil {
			time.Sleep(time.Second)
			continue
		}

		// Start consuming
		msgs, err := ch.Consume(
			queue.Name,
			c.Name, // consumer tag (auto-generated)
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			c.client.logger.Error("Consume failed", "error", err)
			time.Sleep(time.Second)
			continue
		}

		c.processMessages(msgs)
	}
}

func (c *AMQPConsumer) processMessages(msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-c.done:
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			c.deliveryCh <- msg
		}
	}
}

func (c *AMQPConsumer) processMessage(msg amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%+v", r)
			c.client.logger.Error("Recovered from panic", "error", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	eventName := domain.EventName(msg.Type)
	if err := c.Dispatcher.Dispatch(ctx, eventName, msg.Body); err != nil {
		c.client.logger.Error("Process message failed", "error", err)

		requeue := ports.IsErrRequeueable(err)
		c.nack(msg, requeue)
	} else {
		c.ack(msg)
	}
}

func (c *AMQPConsumer) ack(msg amqp.Delivery) {
	t := time.NewTimer(c.AckTimeout)
	for {
		select {
		case <-t.C:
			return
		default:
		}
		if err := msg.Ack(false); err != nil {
			c.client.logger.Error("Ack message failed", "error", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
	}
}

func (c *AMQPConsumer) nack(msg amqp.Delivery, requeue bool) {
	t := time.NewTimer(c.AckTimeout)
	for {
		select {
		case <-t.C:
			return
		default:
		}
		if err := msg.Nack(false, requeue); err != nil {
			c.client.logger.Error("Nack message failed", "error", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
	}
}
