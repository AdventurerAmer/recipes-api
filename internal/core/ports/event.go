package ports

import (
	"context"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type EventPublisher interface {
	Publish(ctx context.Context, event domain.Event) error
}
