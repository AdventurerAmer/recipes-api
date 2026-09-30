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
	Register(ctx context.Context, req RegisterRequest) (RegisterResponse, *domain.User, error)
	Get(ctx context.Context, req GetUserRequest) (GetUserResponse, error)
	SendVerification(ctx context.Context, req SendVerificationRequest) (SendVerificationResponse, error)
	Verify(ctx context.Context, req VerifyUserRequest) (VerifyResponse, error)
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (ForgotPasswordResponse, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) (ResetPasswordResponse, error)
	ChangePassword(ctx context.Context, req ChangePasswordRequest) (ChangePasswordResponse, error)
}

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"displayName" validate:"required,min=8,max=32"`
	Password    string `json:"password" validate:"required,strong_password"`
}

type RegisterResponse struct {
	User    domain.FrontendUser `json:"user"`
	Message string              `json:"message"`
}

type GetUserRequest struct {
	Id string `json:"id"`
}

type GetUserResponse struct {
	User *domain.FrontendUser `json:"user"`
}

type VerifyUserRequest struct {
	Token string `json:"token" form:"token" validate:"required,len=43"`
}

type VerifyResponse struct {
	Message string `json:"message"`
}

type SendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type SendVerificationResponse struct {
	Message string `json:"message"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required,len=43"`
	Password string `json:"password" validate:"required,strong_password"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

type ChangePasswordRequest struct {
	UserId          string `json:"userId" validate:"required"`
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"strong_password"`
}

type ChangePasswordResponse struct {
	Message string `json:"message"`
}
