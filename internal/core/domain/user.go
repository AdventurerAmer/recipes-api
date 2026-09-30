package domain

import "time"

type User struct {
	Id           string    `json:"id" bson:"_id,omitempty"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	Email        string    `json:"email" bson:"email"`
	DisplayName  string    `json:"displayName" bson:"displayName"`
	PasswordHash string    `json:"passwordHash" bson:"passwordHash"`
	IsVerified   bool      `json:"isVerified" bson:"isVerified"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
	Version      int       `json:"version" bson:"version"`
}

type Verification struct {
	Token     string    `json:"token" bson:"token"`
	ExpiresAt time.Time `json:"expiresAt" bson:"expiresAt"`
}

type ForgotPassword struct {
	Token     string    `json:"token" bson:"token"`
	ExpiresAt time.Time `json:"expiresAt" bson:"expiresAt"`
}

type FrontendUser struct {
	Id          string    `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Version     int       `json:"version"`
}

func NewFrontendUser(u *User) FrontendUser {
	return FrontendUser{
		Id:          u.Id,
		CreatedAt:   u.CreatedAt,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		UpdatedAt:   u.UpdatedAt,
		Version:     u.Version,
	}
}
