package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type UserCreated struct {
}

func NewUserCreated() *UserCreated {
	return &UserCreated{}
}

func (h *UserCreated) Name() domain.EventName {
	return domain.EventNameUserCreated
}

func (h *UserCreated) Unmarshal(data []byte) (domain.Event, error) {
	var e domain.UserCreatedEvent
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("'json.Unmarshal' failed: %w", err)
	}
	return &e, nil
}

func (h *UserCreated) Handle(ctx context.Context, event domain.Event) error {
	e := event.(*domain.UserCreatedEvent)
	slog.Info("recived user created event", "user.Id", e.UserId)
	return nil
}
