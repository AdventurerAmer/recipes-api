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
	GetId() string
	SetId(id string)
	GetName() EventName
	SetName(name EventName)
	GetOccurredAt() time.Time
	SetOccurredAt(occurredAt time.Time)
}

type BaseEvent struct {
	Id         string
	Name       EventName
	OccurredAt time.Time
}

func NewBaseEvent(id string, name EventName, occurredAt time.Time) BaseEvent {
	if id == "" {
		id = uuid.NewString()
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	return BaseEvent{
		Id:         id,
		Name:       name,
		OccurredAt: occurredAt,
	}
}

func (e BaseEvent) GetId() string {
	return e.Id
}

func (e *BaseEvent) SetId(id string) {
	e.Id = id
}

func (e BaseEvent) GetName() EventName {
	return e.Name
}

func (e *BaseEvent) SetName(name EventName) {
	e.Name = name
}

func (e BaseEvent) GetOccurredAt() time.Time {
	return e.OccurredAt
}

func (e *BaseEvent) SetOccurredAt(occurredAt time.Time) {
	e.OccurredAt = occurredAt
}
