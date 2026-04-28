package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type redisUsersRepository struct {
	next    ports.UsersRepository
	wrapper redisWrapper
}

func NewRedisUsersRepository(next ports.UsersRepository, client *redis.Client, ttl time.Duration) ports.UsersRepository {
	return &redisUsersRepository{
		next: next,
		wrapper: redisWrapper{
			client: client,
			ttl:    ttl,
		},
	}
}

func (r *redisUsersRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.next.Create(ctx, user); err != nil {
		return fmt.Errorf("'next.Create' failed: %w", err)
	}
	return nil
}

func (r *redisUsersRepository) Get(ctx context.Context, id string) (domain.User, error) {
	key := composeUserByIDKey(id)
	var user domain.User
	cacheErr := r.wrapper.get(ctx, key, &user)
	if cacheErr == nil {
		return user, nil
	}
	user, err := r.next.Get(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("'next.Get' failed: %w", err)
	}
	if cacheErr == redis.Nil {
		_ = r.wrapper.set(ctx, key, user)
	}
	return user, nil
}

func (r *redisUsersRepository) GetByName(ctx context.Context, username string) (domain.User, error) {
	key := composeUserByNameKey(username)
	var user domain.User
	cacheErr := r.wrapper.get(ctx, key, user)
	if cacheErr == nil {
		return user, nil
	}
	user, err := r.next.GetByName(ctx, username)
	if err != nil {
		return domain.User{}, fmt.Errorf("'next.GetByName' failed: %w", err)
	}
	if cacheErr == redis.Nil {
		_ = r.wrapper.set(ctx, key, user)
	}
	return user, nil
}

func (r *redisUsersRepository) Update(ctx context.Context, user *domain.User) error {
	if err := r.next.Update(ctx, user); err != nil {
		return fmt.Errorf("'next.Update' failed: %w", err)
	}
	_ = r.invalidateCache(ctx, *user)
	return nil
}

func (r *redisUsersRepository) Delete(ctx context.Context, user domain.User) error {
	if err := r.next.Delete(ctx, user); err != nil {
		return fmt.Errorf("'next.Update' failed: %w", err)
	}
	_ = r.invalidateCache(ctx, user)
	return nil
}

func (r *redisUsersRepository) invalidateCache(ctx context.Context, user domain.User) error {
	keys := []string{composeUserByIDKey(user.ID), composeUserByNameKey(user.Username)}
	if err := r.wrapper.del(ctx, keys...); err != nil {
		return fmt.Errorf("'wrapper.del' failed: %w", err)
	}
	return nil
}
