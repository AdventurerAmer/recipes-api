package domain

import "time"

type TokenType string

const (
	TokenTypeVerification  TokenType = "verification"
	TokenTypePasswordReset TokenType = "passwordReset"
)

type Token struct {
	Id        string     `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time  `json:"createdAt" bson:"createdAt"`
	Type      TokenType  `json:"type" bson:"type"`
	UserId    string     `json:"userId" bson:"userId"`
	Hash      string     `json:"hash" bson:"hash"`
	ExpiresAt time.Time  `json:"expiresAt" bson:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt" bson:"usedAt"`
}
