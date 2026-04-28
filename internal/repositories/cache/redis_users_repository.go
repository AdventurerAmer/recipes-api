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

type redisUsersRepository struct {
	next   ports.UsersRepository
	client *redis.Client
	ttl    time.Duration
}

func NewRedisUsersRepository(next ports.UsersRepository, client *redis.Client, ttl time.Duration) ports.UsersRepository {
	return &redisUsersRepository{
		next:   next,
		client: client,
		ttl:    ttl,
	}
}

func (r *redisUsersRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.next.Create(ctx, user); err != nil {
		return fmt.Errorf("'next.Create' failed: %w", err)
	}
	return nil
}

func (r *redisUsersRepository) Get(ctx context.Context, id string) (domain.User, error) {
	key := fmt.Sprintf("user id:%s", id)
	data, cacheErr := r.client.Get(ctx, key).Result()
	if cacheErr == nil {
		var user domain.User
		if err := json.Unmarshal([]byte(data), &user); err == nil {
			return user, nil
		}
	}
	user, err := r.next.Get(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("'next.Get' failed: %w", err)
	}
	if cacheErr == redis.Nil {
		if data, err := json.Marshal(user); err != nil {
			_, _ = r.client.Set(ctx, key, data, r.ttl).Result()
		}
	}
	return user, nil
}

func (r *redisUsersRepository) GetByName(ctx context.Context, username string) (domain.User, error) {
	key := fmt.Sprintf("user name:%s", username)
	data, cacheErr := r.client.Get(ctx, key).Result()
	if cacheErr == nil {
		var user domain.User
		if err := json.Unmarshal([]byte(data), &user); err == nil {
			return user, nil
		}
	}
	user, err := r.next.GetByName(ctx, username)
	if err != nil {
		return domain.User{}, fmt.Errorf("'next.GetByName' failed: %w", err)
	}
	if cacheErr == redis.Nil {
		if data, err := json.Marshal(user); err != nil {
			_, _ = r.client.Set(ctx, key, data, r.ttl).Result()
		}
	}
	return user, nil
}

func (r *redisUsersRepository) Update(ctx context.Context, user *domain.User) error {
	if err := r.next.Update(ctx, user); err != nil {
		return fmt.Errorf("'next.Update' failed: %w", err)
	}
	keys := []string{fmt.Sprintf("user id:%s", user.ID), fmt.Sprintf("user name:%s", user.Username)}
	_, _ = r.client.Del(ctx, keys...).Result()
	return nil
}

func (r *redisUsersRepository) Delete(ctx context.Context, id string) error {
	if err := r.next.Delete(ctx, id); err != nil {
		return fmt.Errorf("'next.Update' failed: %w", err)
	}
	key := fmt.Sprintf("user id:%s", id)
	_, _ = r.client.Del(ctx, key).Result()
	return nil
}
