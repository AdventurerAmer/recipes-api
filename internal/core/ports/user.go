package ports

import (
	"context"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type UsersRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetById(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByVerificationToken(ctx context.Context, token string) (*domain.User, error)
	GetByForgotPasswordToken(ctx context.Context, token string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, user *domain.User) error
}

type UsersService interface {
	Register(ctx context.Context, req RegisterRequest) (RegisterResponse, *domain.User, error)
	Verify(ctx context.Context, req VerifyRequest) (VerifyResponse, error)
	SendVerification(ctx context.Context, req SendVerificationRequest) (SendVerificationResponse, error)
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (ForgotPasswordResponse, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) (ResetPasswordResponse, error)
}

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"displayName" validate:"required,min=8,max=32"`
	Password    string `json:"password" validate:"required,strong_password"`
}

type RegisterResponse struct {
	User domain.FrontendUser `json:"user"`
}

type VerifyRequest struct {
	Token string `json:"token" validate:"required,len=26"`
}

type VerifyResponse struct {
}

type SendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type SendVerificationResponse struct {
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required,len=26"`
	Password string `json:"password" validate:"required,strong_password"`
}

type ResetPasswordResponse struct {
}
