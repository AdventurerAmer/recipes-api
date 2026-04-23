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
