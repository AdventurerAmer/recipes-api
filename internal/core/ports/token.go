package ports

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/errs"
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

type TokenManager struct {
	Tokener      Tokener
	Repository   TokensRepository
	Type         domain.TokenType
	ExpiresAfter time.Duration
}

func NewTokenManager(tokener Tokener, repository TokensRepository, tokenType domain.TokenType, expiresAfter time.Duration) *TokenManager {
	return &TokenManager{
		Tokener:      tokener,
		Repository:   repository,
		Type:         tokenType,
		ExpiresAfter: expiresAfter,
	}
}

func (tm *TokenManager) Create(ctx context.Context, userId string) (*domain.Token, string, error) {
	plain, hash, err := tm.Tokener.Generate()
	if err != nil {
		return nil, "", fmt.Errorf("'Tokener.Generate' failed: %w", err)
	}
	now := time.Now().UTC()
	expiresAt := now.Add(tm.ExpiresAfter)
	token := &domain.Token{
		CreatedAt: now,
		Type:      tm.Type,
		UserId:    userId,
		ExpiresAt: expiresAt,
		Hash:      hash,
	}
	if err := tm.Repository.Create(ctx, token); err != nil {
		return nil, "", fmt.Errorf("'Repository.Create' failed: %w", err)
	}
	return token, plain, nil
}

func (tm *TokenManager) Check(ctx context.Context, plain string) (bool, *domain.Token, error) {
	hash := tm.Tokener.Hash(plain)
	token, err := tm.Repository.Get(ctx, tm.Type, hash)
	if err != nil {
		if errs.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("'Repository.Get' failed: %w", err)
	}

	if token.UsedAt != nil {
		return false, nil, nil
	}

	now := time.Now().UTC()
	if now.After(token.ExpiresAt) {
		return false, nil, nil
	}

	return true, token, nil
}

func (tm *TokenManager) Invalidate(ctx context.Context, token *domain.Token) error {
	now := time.Now().UTC()
	token.UsedAt = &now
	if err := tm.Repository.Update(ctx, token); err != nil {
		return fmt.Errorf("'Repository.Update' failed: %w", err)
	}
	return nil
}

func (tm *TokenManager) IsLastTokenStillValid(ctx context.Context, userId string) (bool, error) {
	lastToken, err := tm.Repository.GetLast(ctx, userId, tm.Type)
	if err != nil {
		if errs.IsNotFound(err) {
			return false, nil
		} else {
			return false, fmt.Errorf("'Repository.GetLast' failed: %w", err)
		}
	}

	now := time.Now().UTC()
	if now.After(lastToken.ExpiresAt) {
		return false, nil
	}

	return true, nil
}
