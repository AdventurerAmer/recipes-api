package ports

import (
	"context"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type UsersRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, id string) (domain.User, error)
	GetByName(ctx context.Context, username string) (domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

type UsersService interface {
	SignUp(ctx context.Context, req SignUpRequest) (SignUpResponse, error)
}

type SignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignUpResponse struct {
	User    domain.User `json:"user"`
	Message string      `json:"message"`
}
