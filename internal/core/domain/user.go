package domain

import "time"

// swagger:parameters users newUser
type User struct {
	//swagger:ignore
	ID        string    `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	Username  string    `json:"username" bson:"username"`
	Password  string    `json:"password" bson:"password"`
	Version   int       `json:"version" bson:"version"`
}

func (u User) Frontend() FrontendUser {
	return FrontendUser{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		Username:  u.Username,
		Version:   u.Version,
	}
}

type FrontendUser struct {
	//swagger:ignore
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Username  string    `json:"username"`
	Version   int       `json:"version"`
}
