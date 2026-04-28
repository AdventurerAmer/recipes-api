package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type redisRecipesRepository struct {
	next    ports.RecipesRepository
	wrapper redisWrapper
}

func NewRedisRecipesRepository(next ports.RecipesRepository, client *redis.Client, ttl time.Duration) ports.RecipesRepository {
	return &redisRecipesRepository{
		next: next,
		wrapper: redisWrapper{
			client: client,
			ttl:    ttl,
		},
	}
}

func (r *redisRecipesRepository) Create(ctx context.Context, recipe *domain.Recipe) error {
	if err := r.next.Create(ctx, recipe); err != nil {
		return fmt.Errorf("'next.Create' failed: %w", err)
	}
	_ = r.wrapper.del(ctx, recipesKeyPrefix)
	return nil
}

func (r *redisRecipesRepository) Get(ctx context.Context, id string) (domain.Recipe, error) {
	key := composeRecipeKey(id)

	var recipe domain.Recipe
	cacheErr := r.wrapper.get(ctx, key, &recipe)
	if cacheErr == nil {
		return recipe, nil
	}
	recipe, err := r.next.Get(ctx, id)
	if err != nil {
		return domain.Recipe{}, fmt.Errorf("'next.Get' failed: %w", err)
	}
	if cacheErr == redis.Nil {
		_ = r.wrapper.set(ctx, key, recipe)
	}
	return recipe, nil
}

func (r *redisRecipesRepository) List(ctx context.Context, lastID, userID, sort string, limit int) ([]domain.Recipe, int, error) {
	key := composeRecipesKey(lastID, userID, sort, limit)
	var cachedEntry struct {
		Recipes []domain.Recipe `json:"recipes"`
		Total   int             `json:"total"`
	}
	cacheErr := r.wrapper.get(ctx, key, &cachedEntry)
	if cacheErr == nil {
		return cachedEntry.Recipes, cachedEntry.Total, nil
	}
	recipes, total, err := r.next.List(ctx, lastID, userID, sort, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("'next.List' failed: %w", err)
	}
	if cacheErr == redis.Nil {
		cachedEntry.Recipes = recipes
		cachedEntry.Total = total
		_ = r.wrapper.set(ctx, key, cachedEntry)
	}
	return recipes, total, nil
}

func (r *redisRecipesRepository) Update(ctx context.Context, recipe *domain.Recipe) error {
	if err := r.next.Update(ctx, recipe); err != nil {
		return fmt.Errorf("'next.Update' failed: %w", err)
	}
	_ = r.invalidateCache(ctx, recipe.ID)
	return nil
}

func (r *redisRecipesRepository) Delete(ctx context.Context, userID, id string) error {
	if err := r.next.Delete(ctx, userID, id); err != nil {
		return fmt.Errorf("'next.Update' failed: %w", err)
	}
	_ = r.invalidateCache(ctx, id)
	return nil
}

func (r *redisRecipesRepository) invalidateCache(ctx context.Context, id string) error {
	keys := []string{composeRecipeKey(id), recipesKeyPrefix}
	err := r.wrapper.del(ctx, keys...)
	if err != nil {
		return fmt.Errorf("'wrapper.del' failed: %w", err)
	}
	return nil
}
