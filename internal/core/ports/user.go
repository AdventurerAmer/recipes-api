package ports

import (
	"context"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type UsersRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetById(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, user *domain.User) error
}

type UsersService interface {
	SignUp(ctx context.Context, req SignUpRequest) (SignUpResponse, *domain.User, error)
	SignIn(ctx context.Context, req SignInRequest) (SignInResponse, error)
}

type SignUpRequest struct {
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"displayName" validate:"required,min=8,max=32"`
	Password    string `json:"password" validate:"required,strong_password"`
}

type SignUpResponse struct {
	User domain.FrontendUser `json:"user"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,strong_password"`
}

type SignInResponse struct {
	User *domain.User `json:"user"`
}
