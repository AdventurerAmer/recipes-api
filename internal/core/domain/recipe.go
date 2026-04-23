package domain

import (
	"time"
)

// swagger:parameters recipes newRecipe
type Recipe struct {
	//swagger:ignore
	ID           string    `json:"id" bson:"_id,omitempty"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UserID       string    `json:"userID" bson:"userID"`
	Name         string    `json:"name" bson:"name"`
	Tags         []string  `json:"tags" bson:"tags"`
	Ingredients  []string  `json:"ingredients" bson:"ingredients"`
	Instructions []string  `json:"instructions" bson:"instructions"`
	Version      int       `json:"version" bson:"version"`
}
