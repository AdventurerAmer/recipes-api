package domain

type UserCreatedEvent struct {
	BaseEvent
	UserId string `json:"userId"`
}

func NewUserCreated(userId string) UserCreatedEvent {
	return UserCreatedEvent{
		BaseEvent: NewBaseEvent(EventNameUserCreated),
		UserId:    userId,
	}
}

type UserVerificationEvent struct {
	BaseEvent
	UserId string `json:"userId"`
}

func NewUserVerification(userId string) UserVerificationEvent {
	return UserVerificationEvent{
		BaseEvent: NewBaseEvent(EventNameUserVerification),
		UserId:    userId,
	}
}

type UserPasswordResetEvent struct {
	BaseEvent
	UserId string `json:"userId"`
}

func NewUserPasswordReset(userId string) UserPasswordResetEvent {
	return UserPasswordResetEvent{
		BaseEvent: NewBaseEvent(EventNameUserPasswordReset),
		UserId:    userId,
	}
}
