package userssrv

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/AdventurerAmer/recipes-api/tokens"
	"github.com/AdventurerAmer/recipes-api/validation"
)

type Config struct {
	VerificationTokenLength         int
	VerificationTokenExpiresAfter   time.Duration
	ForgotpasswordTokenLength       int
	ForgotpasswordTokenExpiresAfter time.Duration
	PasswordHasher                  ports.PasswordHasher
	UsersRepo                       ports.UsersRepository
}

type service struct {
	Config
}

func New(cfg Config) ports.UsersService {
	if cfg.VerificationTokenLength == 0 {
		cfg.VerificationTokenLength = 26
	}
	if cfg.VerificationTokenExpiresAfter == 0 {
		cfg.VerificationTokenExpiresAfter = 10 * time.Minute
	}
	if cfg.ForgotpasswordTokenLength == 0 {
		cfg.ForgotpasswordTokenLength = 26
	}
	if cfg.ForgotpasswordTokenExpiresAfter == 0 {
		cfg.ForgotpasswordTokenExpiresAfter = 10 * time.Minute
	}
	return &service{
		Config: cfg,
	}
}

func (srv *service) Register(ctx context.Context, req ports.RegisterRequest) (ports.RegisterResponse, *domain.User, error) {
	if err := validation.Validate(req); err != nil {
		return ports.RegisterResponse{}, nil, fmt.Errorf("validation failed: %w", err)
	}

	hash, err := srv.PasswordHasher.Hash(req.Password)
	if err != nil {
		return ports.RegisterResponse{}, nil, fmt.Errorf("'hashPassward' failed: %w", err)
	}

	token, err := tokens.CryptoBase64(srv.VerificationTokenLength)
	if err != nil {
		return ports.RegisterResponse{}, nil, fmt.Errorf("'tokens.CryptoBase64' failed: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		CreatedAt:   now,
		UpdatedAt:   now,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Verification: domain.Verification{
			Token:     token,
			ExpiresAt: now.Add(srv.VerificationTokenExpiresAfter),
		},
		PasswordHash: hash,
	}
	if err := srv.UsersRepo.Create(ctx, user); err != nil {
		return ports.RegisterResponse{}, nil, fmt.Errorf("'UsersRepo.Create' failed: %w", err)
	}

	// TODO: send email here

	resp := ports.RegisterResponse{
		User: domain.NewFrontendUser(user),
	}
	return resp, user, nil
}

func (srv *service) Verify(ctx context.Context, req ports.VerifyRequest) (ports.VerifyResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.VerifyResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.GetByVerificationToken(ctx, req.Token)
	if err != nil {
		return ports.VerifyResponse{}, fmt.Errorf("'UsersRepo.GetByVerificationToken' failed: %w", err)
	}

	now := time.Now().UTC()
	if now.After(user.Verification.ExpiresAt) {
		return ports.VerifyResponse{}, fmt.Errorf("invalid or expired token")
	}

	user.IsVerified = true
	user.Verification = domain.Verification{}
	user.UpdatedAt = now

	if err := srv.UsersRepo.Update(ctx, user); err != nil {
		return ports.VerifyResponse{}, fmt.Errorf("'UsersRepo.Update' failed: %w", err)
	}

	return ports.VerifyResponse{}, nil
}

func (srv *service) SendVerification(ctx context.Context, req ports.SendVerificationRequest) (ports.SendVerificationResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("validation failed: %w", err)
	}
	user, err := srv.UsersRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("'UsersRepo.GetByEmail' failed: %w", err)
	}

	now := time.Now().UTC()
	if user.Verification.ExpiresAt.After(now) {
		return ports.SendVerificationResponse{}, fmt.Errorf("email already sent")
	}

	token, err := tokens.CryptoBase64(srv.VerificationTokenLength)
	if err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("'tokens.CryptoBase64' failed: %w", err)
	}

	user.Verification.Token = token
	user.Verification.ExpiresAt = now.Add(srv.VerificationTokenExpiresAfter)
	user.UpdatedAt = now

	if err := srv.UsersRepo.Update(ctx, user); err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("'UsersRepo.Update' failed: %w", err)
	}

	// TODO: send email here

	return ports.SendVerificationResponse{}, nil
}

func (srv *service) ForgotPassword(ctx context.Context, req ports.ForgotPasswordRequest) (ports.ForgotPasswordResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("'UsersRepo.GetByEmail' failed: %w", err)
	}

	now := time.Now().UTC()
	if user.Verification.ExpiresAt.After(now) {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("email already sent")
	}

	token, err := tokens.CryptoBase64(srv.ForgotpasswordTokenLength)
	if err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("'tokens.CryptoBase64' failed: %w", err)
	}

	user.ForgotPassword.Token = token
	user.ForgotPassword.ExpiresAt = now.Add(srv.ForgotpasswordTokenExpiresAfter)
	user.UpdatedAt = now
	if err := srv.UsersRepo.Update(ctx, user); err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("'UsersRepo.Update' failed: %w", err)
	}

	// TODO: send email here

	return ports.ForgotPasswordResponse{}, nil
}

func (srv *service) ResetPassword(ctx context.Context, req ports.ResetPasswordRequest) (ports.ResetPasswordResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.GetByForgotPasswordToken(ctx, req.Token)
	if err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("'UsersRepo.GetByVerificationToken' failed: %w", err)
	}

	now := time.Now().UTC()
	if now.After(user.ForgotPassword.ExpiresAt) {
		return ports.ResetPasswordResponse{}, fmt.Errorf("invalid or expired token")
	}

	hash, err := srv.PasswordHasher.Hash(req.Password)
	if err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("'PasswordHasher.Hash' failed: %w", err)
	}

	user.PasswordHash = hash
	user.ForgotPassword = domain.ForgotPassword{}
	user.UpdatedAt = now

	if err := srv.UsersRepo.Update(ctx, user); err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("'UsersRepo.Update' failed: %w", err)
	}

	return ports.ResetPasswordResponse{}, nil
}
