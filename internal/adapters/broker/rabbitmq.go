package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPClient struct {
	connStr       string
	conn          *amqp.Connection
	publishCh     *amqp.Channel
	consumeCh     *amqp.Channel
	done          chan struct{}
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	isConnected   bool
	mu            sync.RWMutex
	publishMu     sync.Mutex
}

func NewAMQPClient(connStr string) (*AMQPClient, error) {
	client := &AMQPClient{
		connStr: connStr,
		done:    make(chan struct{}),
	}
	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("'client.connect' failed: %w", err)
	}

	go client.handleReconnect()

	return client, nil
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

	// Enable publisher confirms for reliable publishing
	if err := publishCh.Confirm(false); err != nil {
		_ = consumeCh.Close()
		_ = publishCh.Close()
		_ = conn.Close()
		return fmt.Errorf("'publishCh.Confirm' failed: %w", err)
	}

	for _, eventName := range domain.EventNames {
		exchange := eventName.String()
		durable := true
		if err := publishCh.ExchangeDeclare(exchange, "fanout", durable, false, false, false, nil); err != nil {
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
	c.conn = conn
	c.publishCh = publishCh
	c.consumeCh = consumeCh
	c.notifyClose = notifyClose
	c.notifyConfirm = notifyConfirm
	c.isConnected = true
	c.mu.Unlock()

	return nil
}

func (c *AMQPClient) handleReconnect() {
	for {
		select {
		case <-c.done:
			return
		case err := <-c.notifyClose:
			if err != nil {
				// TODO: log here
			}

			c.mu.Lock()
			c.isConnected = false
			c.mu.Unlock()

			// Reconnect with exponential backoff
			c.reconnectWithBackoff()
		}
	}
}

// reconnectWithBackoff attempts to reconnect with increasing delays
func (c *AMQPClient) reconnectWithBackoff() {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-c.done:
			return
		default:
		}

		// log.Printf("Attempting to reconnect in %v...", backoff)
		time.Sleep(backoff)

		if err := c.connect(); err != nil {
			// log.Printf("Failed to reconnect: %v", err)
			// Exponential backoff with cap
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// Successfully reconnected, reset backoff
		// log.Println("Reconnected successfully")
		return
	}
}

func (c *AMQPClient) Publish(ctx context.Context, event domain.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}

	c.publishMu.Lock()
	defer c.publishMu.Unlock()

	for {
		// Check if we're connected
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
			// Connection might have dropped, wait and retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}

		// Wait for confirmation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case confirm, ok := <-confirms:
			if !ok {
				// Channel closed before the publish was confirmed
				continue
			}
			if confirm.Ack {
				return nil
			}
			// log.Println("Message was nacked, retrying...")
			continue
		}
	}
}

// Close gracefully shuts down the client
func (c *AMQPClient) Close() error {
	close(c.done)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.publishCh != nil {
		if err := c.publishCh.Close(); err != nil {
			// log.Printf("Error closing publish channel: %v", err)
		}
	}

	if c.consumeCh != nil {
		if err := c.consumeCh.Close(); err != nil {
			// log.Printf("Error closing consume channel: %v", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("'conn.Close' failed: %w", err)
		}
	}

	c.isConnected = false

	// log.Println("Client closed")

	return nil
}

type AMQPConsumer struct {
	name     string
	client   *AMQPClient
	registry ports.EventRegistry
	done     chan struct{}
	wg       sync.WaitGroup
}

func NewAMQPConsumer(name string, client *AMQPClient, registry ports.EventRegistry) *AMQPConsumer {
	return &AMQPConsumer{
		name:     name,
		client:   client,
		registry: registry,
		done:     make(chan struct{}),
	}
}

// Start begins consuming messages, automatically recovering from disconnections
func (c *AMQPConsumer) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consumeLoop()
	}()
}

func (c *AMQPConsumer) consumeLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
		}

		// Wait for connection
		c.client.mu.RLock()
		connected := c.client.isConnected
		ch := c.client.consumeCh
		c.client.mu.RUnlock()

		if !connected || ch == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		eventNames := slices.Collect(maps.Keys(c.registry))

		durable := true
		queue, err := ch.QueueDeclare(
			c.name,  // name
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
			exchange := eventName.String()
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
			c.name, // consumer tag (auto-generated)
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			log.Printf("Failed to start consuming: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// log.Printf("Started consuming from %s", c.queue)

		// Process messages until disconnection
		c.processMessages(msgs)

		// log.Println("Consumer disconnected, will retry...")
	}
}

func (c *AMQPConsumer) processMessages(msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-c.done:
			return
		case msg, ok := <-msgs:
			if !ok {
				// Channel closed, need to reconnect
				return
			}

			eventName := domain.EventName(msg.Type)
			slog.Info("got event", "type", eventName)
			handler := c.registry[eventName]
			event, err := handler.Unmarshal(msg.Body)
			if err != nil {
				log.Printf("Error processing message: %v", err)
				_ = msg.Nack(false, true) // requeue
				continue
			}

			if err := handler.Handle(context.Background(), event); err != nil {
				log.Printf("Error processing message: %v", err)
				_ = msg.Nack(false, true) // requeue
				continue
			}

			_ = msg.Ack(false)
		}
	}
}

// Stop gracefully stops the consumer
func (c *AMQPConsumer) Stop() {
	close(c.done)
	c.wg.Wait()
}
