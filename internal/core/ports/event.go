package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

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

type EventDispatcher struct {
	handlers map[string]func(context.Context, json.RawMessage) error
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string]func(context.Context, json.RawMessage) error),
	}
}

// Register connects an event type name to a type-safe handler function.
func RegisterEvent[E domain.Event](d *EventDispatcher, eventName domain.EventName, handler EventHandler[E]) {

	// Create a non-generic wrapper that hides the type casting internally
	d.handlers[eventName.String()] = func(ctx context.Context, raw json.RawMessage) error {
		var payload E
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload for event %s: %w", eventName, err)
		}
		return handler(ctx, payload)
	}
}

func (d *EventDispatcher) Dispatch(ctx context.Context, eventName domain.EventName, raw json.RawMessage) error {
	handler, exists := d.handlers[eventName.String()]
	if !exists {
		return fmt.Errorf("no handler registered for event: %s", eventName)
	}

	return handler(ctx, raw)
}

func (d *EventDispatcher) Events() []string {
	return slices.Collect(maps.Keys(d.handlers))
}
