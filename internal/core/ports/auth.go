package ports

import (
	"context"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	Logout(ctx context.Context, req LogoutRequest) (LogoutResponse, error)
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,strong_password"`
}

type LoginResponse struct {
	User *domain.FrontendUser `json:"user"`
}

type LogoutRequest struct {
	Message string `json:"message"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}
