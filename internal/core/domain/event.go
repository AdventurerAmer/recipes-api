package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventName string

const (
	EventNameUserCreated       = "user.created"
	EventNameUserVerification  = "user.verification"
	EventNameUserPasswordReset = "user.passwordReset"
)

func (e EventName) String() string {
	return string(e)
}

var EventNames = []EventName{
	EventNameUserCreated,
	EventNameUserVerification,
	EventNameUserPasswordReset,
}

type Event interface {
	Id() string
	Name() EventName
	OccurredAt() time.Time
}

type BaseEvent struct {
	id         string
	name       EventName
	occurredAt time.Time
}

func NewBaseEvent(id string, name EventName, occurredAt time.Time) BaseEvent {
	if id == "" {
		id = uuid.NewString()
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	return BaseEvent{
		id:         id,
		name:       name,
		occurredAt: occurredAt,
	}
}

func (e BaseEvent) Id() string {
	return e.id
}

func (e BaseEvent) Name() EventName {
	return e.name
}

func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}
