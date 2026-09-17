package domain

import "time"

type User struct {
	Id           string    `json:"id" bson:"_id,omitempty"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	Email        string    `json:"email" bson:"email"`
	DisplayName  string    `json:"displayName" bson:"displayName"`
	PasswordHash string    `json:"passwordHash" bson:"passwordHash"`
	Version      int       `json:"version" bson:"version"`
}

type FrontendUser struct {
	Id          string    `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Version     int       `json:"version"`
}

func NewFrontendUser(u *User) FrontendUser {
	return FrontendUser{
		Id:          u.Id,
		CreatedAt:   u.CreatedAt,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Version:     u.Version,
	}
}
