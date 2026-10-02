package ports

import (
	"context"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type TokensRepository interface {
	Create(ctx context.Context, token *domain.Token) error
	Get(ctx context.Context, tokenType domain.TokenType, hash string) (*domain.Token, error)
	GetLast(ctx context.Context, userId string, tokenType domain.TokenType) (*domain.Token, error)
	Update(ctx context.Context, token *domain.Token) error
	Delete(ctx context.Context, token *domain.Token) error
}

type TokenGenerator interface {
	Generate() (plain string, hash string, err error)
}

type TokenHasher interface {
	Hash(plain string) string
}

type TokenVerifier interface {
	Verify(plain, hash string) bool
}

type Tokener interface {
	TokenGenerator
	TokenHasher
	TokenVerifier
}
