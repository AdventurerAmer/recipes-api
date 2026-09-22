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

func (e BaseEvent) Id() string {
	return e.id
}

func (e BaseEvent) Name() EventName {
	return e.name
}

func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

type UserCreatedEvent struct {
	BaseEvent
	UserId      string
	Email       string
	DisplayName string
}

func NewUserCreated(userId, email, displayName string) UserCreatedEvent {
	return UserCreatedEvent{
		BaseEvent: BaseEvent{
			id:         uuid.NewString(),
			name:       EventNameUserCreated,
			occurredAt: time.Now().UTC(),
		},
		UserId:      userId,
		Email:       email,
		DisplayName: displayName,
	}
}

type UserVerificationEvent struct {
	BaseEvent
	UserId      string
	Email       string
	DisplayName string
}

func NewUserVerification(userId, email, displayName string) UserVerificationEvent {
	return UserVerificationEvent{
		BaseEvent: BaseEvent{
			id:         uuid.NewString(),
			name:       EventNameUserVerification,
			occurredAt: time.Now().UTC(),
		},
		UserId:      userId,
		Email:       email,
		DisplayName: displayName,
	}
}

type UserPasswordResetEvent struct {
	BaseEvent
	UserId string
	Email  string
}

func NewUserPasswordReset(userId, email string) UserPasswordResetEvent {
	return UserPasswordResetEvent{
		BaseEvent: BaseEvent{
			id:         uuid.NewString(),
			name:       EventNameUserPasswordReset,
			occurredAt: time.Now().UTC(),
		},
		UserId: userId,
		Email:  email,
	}
}
