package ports

import (
	"context"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type TokensRepository interface {
	Create(ctx context.Context, token *domain.Token) error
	Get(ctx context.Context, tokenType domain.TokenType, hash string) (*domain.Token, error)
	Update(ctx context.Context, token *domain.Token) error
	Delete(ctx context.Context, token *domain.Token) error
}
