package domain

import "time"

type UserCreatedEvent struct {
	BaseEvent
	UserId                     string    `json:"userId" validate:"required"`
	VerificationToken          string    `json:"verificationToken" validate:"required"`
	VerificationTokenExpiresAt time.Time `json:"verificationTokenExpiresAt" validate:"required"`
}

func NewUserCreated(userId, verificationToken string, verificationTokenExpiresAt time.Time) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseEvent:                  NewBaseEvent("", EventNameUserCreated, time.Time{}),
		UserId:                     userId,
		VerificationToken:          verificationToken,
		VerificationTokenExpiresAt: verificationTokenExpiresAt,
	}
}

type UserVerificationEvent struct {
	BaseEvent
	UserId    string    `json:"userId" validate:"required"`
	Token     string    `json:"token" validate:"required"`
	ExpiresAt time.Time `json:"expiresAt" validate:"required"`
}

func NewUserVerification(userId, token string, expiresAt time.Time) *UserVerificationEvent {
	return &UserVerificationEvent{
		BaseEvent: NewBaseEvent("", EventNameUserVerification, time.Time{}),
		UserId:    userId,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}

type UserPasswordResetEvent struct {
	BaseEvent
	UserId    string    `json:"userId" validate:"required"`
	Token     string    `json:"token" validate:"required"`
	ExpiresAt time.Time `json:"expiresAt" validate:"required"`
}

func NewUserPasswordReset(userId, token string, expiresAt time.Time) *UserPasswordResetEvent {
	return &UserPasswordResetEvent{
		BaseEvent: NewBaseEvent("", EventNameUserPasswordReset, time.Time{}),
		UserId:    userId,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}
