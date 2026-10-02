package tokens

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/errs"
	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
)

type Manager struct {
	Tokener      ports.Tokener
	Repository   ports.TokensRepository
	Type         domain.TokenType
	ExpiresAfter time.Duration
}

func NewManager(tokener ports.Tokener, repository ports.TokensRepository, tokenType domain.TokenType, expiresAfter time.Duration) *Manager {
	return &Manager{
		Tokener:      tokener,
		Repository:   repository,
		Type:         tokenType,
		ExpiresAfter: expiresAfter,
	}
}

func (tm *Manager) Create(ctx context.Context, userId string) (*domain.Token, string, error) {
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

func (tm *Manager) Check(ctx context.Context, plain string) (bool, *domain.Token, error) {
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

func (tm *Manager) Invalidate(ctx context.Context, token *domain.Token) error {
	now := time.Now().UTC()
	token.UsedAt = &now
	if err := tm.Repository.Update(ctx, token); err != nil {
		return fmt.Errorf("'Repository.Update' failed: %w", err)
	}
	return nil
}

func (tm *Manager) IsLastTokenStillValid(ctx context.Context, userId string) (bool, error) {
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
