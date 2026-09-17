package domain

import (
	"time"
)

type Recipe struct {
	Id           string    `json:"id" bson:"_id,omitempty"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UserId       string    `json:"userId" bson:"userId"`
	Name         string    `json:"name" bson:"name"`
	Tags         []string  `json:"tags" bson:"tags"`
	Ingredients  []string  `json:"ingredients" bson:"ingredients"`
	Instructions []string  `json:"instructions" bson:"instructions"`
	ImageURL     string    `json:"imageURL" bson:"imageURL"`
	Version      int       `json:"version" bson:"version"`
}
