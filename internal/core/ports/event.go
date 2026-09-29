package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type EventPublisher interface {
	Publish(ctx context.Context, event domain.Event) error
}

type EventSubscriber interface {
	Start()
	Stop()
}

type EventHandler[E domain.Event] func(ctx context.Context, e E) error

type DispatcherFunc func(ctx context.Context, id string, occurredAt time.Time, payload json.RawMessage) error

type EventDispatcher struct {
	handlers map[domain.EventName]DispatcherFunc
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[domain.EventName]DispatcherFunc),
	}
}

// Register connects an event type name to a type-safe handler function.
func RegisterEvent[E domain.Event](d *EventDispatcher, eventName domain.EventName, handler EventHandler[E]) {

	// Create a non-generic wrapper that hides the type casting internally
	d.handlers[eventName] = func(ctx context.Context, id string, occurredAt time.Time, payload json.RawMessage) error {
		var event E
		if err := json.Unmarshal(payload, &event); err != nil {
			return fmt.Errorf("failed to unmarshal payload for event %s: %w", eventName, err)
		}
		event.SetId(id)
		event.SetOccurredAt(occurredAt)
		event.SetName(eventName)

		return handler(ctx, event)
	}
}

func (d *EventDispatcher) Dispatch(ctx context.Context, eventName domain.EventName, id string, occurredAt time.Time, payload json.RawMessage) error {
	handler, exists := d.handlers[eventName]
	if !exists {
		return fmt.Errorf("no handler registered for event: %s", eventName)
	}

	return handler(ctx, id, occurredAt, payload)
}

func (d *EventDispatcher) Events() []domain.EventName {
	return slices.Collect(maps.Keys(d.handlers))
}
