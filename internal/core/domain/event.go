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
	Key() string
	Name() EventName
	OccurredAt() time.Time
}

type BaseEvent struct {
	key        string
	name       EventName
	occurredAt time.Time
}

func NewBaseEvent(name EventName) BaseEvent {
	return BaseEvent{
		key:        uuid.NewString(),
		name:       name,
		occurredAt: time.Now().UTC(),
	}
}

func (e BaseEvent) Key() string {
	return e.key
}

func (e BaseEvent) Name() EventName {
	return e.name
}

func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

type UserCreatedEvent struct {
	BaseEvent
	UserId string
}

func NewUserCreated(userId string) UserCreatedEvent {
	return UserCreatedEvent{
		BaseEvent: NewBaseEvent(EventNameUserCreated),
		UserId:    userId,
	}
}

type UserVerificationEvent struct {
	BaseEvent
	UserId string
}

func NewUserVerification(userId string) UserVerificationEvent {
	return UserVerificationEvent{
		BaseEvent: NewBaseEvent(EventNameUserVerification),
		UserId:    userId,
	}
}

type UserPasswordResetEvent struct {
	BaseEvent
	UserId string
}

func NewUserPasswordReset(userId string) UserPasswordResetEvent {
	return UserPasswordResetEvent{
		BaseEvent: NewBaseEvent(EventNameUserPasswordReset),
		UserId:    userId,
	}
}
