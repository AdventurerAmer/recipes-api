package authsrv

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/errs"
	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
)

type Config struct {
	PasswordVerifier ports.PasswordVerifier
	UsersRepo        ports.UsersRepository
}

type service struct {
	*Config
}

func New(cfg *Config) ports.AuthService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Login(ctx context.Context, req ports.LoginRequest) (ports.LoginResponse, error) {
	user, err := srv.UsersRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return ports.LoginResponse{}, fmt.Errorf("'UsersRepo.GetByEmail' failed: %w", err)
	}
	ok, err := srv.PasswordVerifier.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return ports.LoginResponse{}, fmt.Errorf("'verifyPassword' failed: %w", err)
	}
	if !ok {
		return ports.LoginResponse{}, errs.NewAuthentication("invalid credentials")
	}
	frontendUser := domain.NewFrontendUser(user)
	return ports.LoginResponse{
		User: &frontendUser,
	}, nil
}

func (srv *service) Logout(ctx context.Context, req ports.LogoutRequest) (ports.LogoutResponse, error) {
	return ports.LogoutResponse{}, nil
}
