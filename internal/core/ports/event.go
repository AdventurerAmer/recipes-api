package ports

import (
	"context"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type EventPublisher interface {
	Publish(ctx context.Context, event domain.Event) error
}

type EventSubscriber interface {
	Start()
	Stop()
}

type EventHandler interface {
	Name() domain.EventName
	Unmarshal(data []byte) (domain.Event, error)
	Handle(ctx context.Context, event domain.Event) error
}

type EventRegistry map[domain.EventName]EventHandler

func NewEventRegistry() EventRegistry {
	return make(map[domain.EventName]EventHandler)
}

func (r EventRegistry) Register(handler EventHandler) {
	eventName := handler.Name()
	if _, ok := r[eventName]; ok {
		panic("event already registered")
	}
	r[eventName] = handler
}
