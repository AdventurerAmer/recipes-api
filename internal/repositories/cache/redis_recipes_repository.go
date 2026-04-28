package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type redisRecipesRepository struct {
	next   ports.RecipesRepository
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRecipesRepository(next ports.RecipesRepository, client *redis.Client, ttl time.Duration) ports.RecipesRepository {
	return &redisRecipesRepository{
		next:   next,
		client: client,
		ttl:    ttl,
	}
}

func (r *redisRecipesRepository) Create(ctx context.Context, recipe *domain.Recipe) error {
	if err := r.next.Create(ctx, recipe); err != nil {
		return fmt.Errorf("'next.Create' failed: %w", err)
	}
	r.client.Del(ctx, "recipes:*")
	return nil
}

func (r *redisRecipesRepository) Get(ctx context.Context, id string) (domain.Recipe, error) {
	key := fmt.Sprintf("recipe id:%s", id)
	data, cacheErr := r.client.Get(ctx, key).Result()
	if cacheErr == nil {
		var recipe domain.Recipe
		if err := json.Unmarshal([]byte(data), &recipe); err == nil {
			return recipe, nil
		}
	}
	recipe, err := r.next.Get(ctx, id)
	if err != nil {
		return domain.Recipe{}, fmt.Errorf("'next.Get' failed: %w", err)
	}
	if cacheErr == redis.Nil {
		if data, err := json.Marshal(recipe); err != nil {
			_, _ = r.client.Set(ctx, key, data, r.ttl).Result()
		}
	}
	return recipe, nil
}

func (r *redisRecipesRepository) List(ctx context.Context, lastID, userID, sort string, limit int) ([]domain.Recipe, int, error) {
	key := fmt.Sprintf("recipes: lastID=%s,sort=%s,limit=%d", lastID, sort, limit)
	data, cacheErr := r.client.Get(ctx, key).Result()
	type recipesCacheEntry struct {
		Recipes []domain.Recipe `json:"recipes"`
		Total   int             `json:"total"`
	}
	if cacheErr == nil {
		var entry recipesCacheEntry
		if err := json.Unmarshal([]byte(data), &entry); err == nil {
			return entry.Recipes, entry.Total, nil
		}
	}
	recipes, total, err := r.next.List(ctx, lastID, userID, sort, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("'.Lnextist' failed: %w", err)
	}
	if cacheErr == redis.Nil {
		entry := recipesCacheEntry{
			Recipes: recipes,
			Total:   total,
		}
		if data, err := json.Marshal(entry); err != nil {
			_, _ = r.client.Set(ctx, key, data, r.ttl).Result()
		}
	}
	return recipes, total, nil
}

func (r *redisRecipesRepository) Update(ctx context.Context, recipe *domain.Recipe) error {
	if err := r.next.Update(ctx, recipe); err != nil {
		return fmt.Errorf("'next.Update' failed: %w", err)
	}
	keys := []string{fmt.Sprintf("recipe id:%s", recipe.ID), "recipes:*"}
	_, _ = r.client.Del(ctx, keys...).Result()
	return nil
}

func (r *redisRecipesRepository) Delete(ctx context.Context, userID, id string) error {
	if err := r.next.Delete(ctx, userID, id); err != nil {
		return fmt.Errorf("'next.Update' failed: %w", err)
	}
	keys := []string{fmt.Sprintf("recipe id:%s", id), "recipes:*"}
	_, _ = r.client.Del(ctx, keys...).Result()
	return nil
}
