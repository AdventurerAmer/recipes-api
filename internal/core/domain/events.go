package domain

import "time"

type UserCreatedEvent struct {
	BaseEvent
	UserId string `json:"userId"`
}

func NewUserCreated(userId string) UserCreatedEvent {
	return UserCreatedEvent{
		BaseEvent: NewBaseEvent("", EventNameUserCreated, time.Time{}),
		UserId:    userId,
	}
}

type UserVerificationEvent struct {
	BaseEvent
	UserId string `json:"userId"`
}

func NewUserVerification(userId string) UserVerificationEvent {
	return UserVerificationEvent{
		BaseEvent: NewBaseEvent("", EventNameUserVerification, time.Time{}),
		UserId:    userId,
	}
}

type UserPasswordResetEvent struct {
	BaseEvent
	UserId string `json:"userId"`
}

func NewUserPasswordReset(userId string) UserPasswordResetEvent {
	return UserPasswordResetEvent{
		BaseEvent: NewBaseEvent("", EventNameUserPasswordReset, time.Time{}),
		UserId:    userId,
	}
}
