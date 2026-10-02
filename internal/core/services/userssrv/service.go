package userssrv

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/errs"
	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/AdventurerAmer/recipes-api/tokens"
	"github.com/AdventurerAmer/recipes-api/validation"
)

type Config struct {
	PasswordManager           ports.PasswordManager
	VerificationTokenManager  *tokens.Manager
	PasswordResetTokenManager *tokens.Manager
	Transactor                ports.Transactor
	UsersRepo                 ports.UsersRepository
	EventPublisher            ports.EventPublisher
}

type service struct {
	Config
}

func New(cfg Config) ports.UsersService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Register(ctx context.Context, req ports.RegisterRequest) (ports.RegisterResponse, *domain.User, error) {
	if err := validation.Validate(req); err != nil {
		return ports.RegisterResponse{}, nil, fmt.Errorf("validation failed: %w", err)
	}

	passwordHash, err := srv.PasswordManager.Hash(req.Password)
	if err != nil {
		return ports.RegisterResponse{}, nil, fmt.Errorf("'hashPassward' failed: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		CreatedAt:    now,
		UpdatedAt:    now,
		Email:        req.Email,
		DisplayName:  req.DisplayName,
		PasswordHash: passwordHash,
	}

	txn := func(tctx context.Context) error {
		if err := srv.UsersRepo.Create(tctx, user); err != nil {
			return fmt.Errorf("'UsersRepo.Create' failed: %w", err)
		}

		token, plain, err := srv.VerificationTokenManager.Create(tctx, user.Id)
		if err != nil {
			return fmt.Errorf("'VerificationTokenManager.Create' failed: %w", err)
		}

		event := domain.NewUserCreated(user.Id, plain, token.ExpiresAt)
		if err := srv.EventPublisher.Publish(tctx, event); err != nil {
			return fmt.Errorf("'EventPublisher.Publish' failed: %w", err)
		}

		return nil
	}
	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.RegisterResponse{}, nil, fmt.Errorf("transaction failed: %w", err)
	}

	resp := ports.RegisterResponse{
		User:    domain.NewFrontendUser(user),
		Message: "User was registered successfully",
	}
	return resp, user, nil
}

func (srv *service) Get(ctx context.Context, req ports.GetUserRequest) (ports.GetUserResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.GetUserResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.GetById(ctx, req.Id)
	if err != nil {
		return ports.GetUserResponse{}, fmt.Errorf("'UsersRepo.GetById' failed: %w", err)
	}

	frontendUser := domain.NewFrontendUser(user)
	return ports.GetUserResponse{
		User: &frontendUser,
	}, nil
}

func (srv *service) Update(ctx context.Context, req ports.UpdateUserRequest) (ports.UpdateUserResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.UpdateUserResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.GetById(ctx, req.UserId)
	if err != nil {
		return ports.UpdateUserResponse{}, fmt.Errorf("'UsersRepo.GetById' failed: %w", err)
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}

	if err := srv.UsersRepo.Update(ctx, user); err != nil {
		return ports.UpdateUserResponse{}, fmt.Errorf("'UsersRepo.Update' failed: %w", err)
	}

	frontendUser := domain.NewFrontendUser(user)
	return ports.UpdateUserResponse{
		User: &frontendUser,
	}, nil
}

func (srv *service) Verify(ctx context.Context, req ports.VerifyUserRequest) (ports.VerifyResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.VerifyResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	ok, token, err := srv.VerificationTokenManager.Check(ctx, req.Token)
	if err != nil {
		return ports.VerifyResponse{}, fmt.Errorf("VerificationTokenManager.Check")
	}

	if !ok {
		return ports.VerifyResponse{}, errs.NewFailedPrecondition("invalid or expired token")
	}

	txn := func(tctx context.Context) error {
		user, err := srv.UsersRepo.GetById(tctx, token.UserId)
		if err != nil {
			return fmt.Errorf("'UsersRepo.GetById' failed: %w", err)
		}

		now := time.Now().UTC()
		user.IsVerified = true
		user.UpdatedAt = now
		if err := srv.UsersRepo.Update(tctx, user); err != nil {
			return fmt.Errorf("'UsersRepo.Update' failed: %w", err)
		}

		if err := srv.VerificationTokenManager.Invalidate(ctx, token); err != nil {
			return fmt.Errorf("'VerificationTokenManager.Update' failed: %w", err)
		}

		return nil
	}
	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.VerifyResponse{}, fmt.Errorf("transaction failed: %w", err)
	}

	return ports.VerifyResponse{
		Message: "User was verified successfully",
	}, nil
}

func (srv *service) SendVerification(ctx context.Context, req ports.SendVerificationRequest) (ports.SendVerificationResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("'UsersRepo.GetByEmail' failed: %w", err)
	}

	if user.IsVerified {
		return ports.SendVerificationResponse{}, errs.NewFailedPrecondition("user is already verified")
	}

	ok, err := srv.VerificationTokenManager.IsLastTokenStillValid(ctx, user.Id)
	if err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("'VerificationTokenManager.IsLastTokenStillValid' failed: %w", err)
	}

	if ok {
		return ports.SendVerificationResponse{}, errs.NewFailedPrecondition("email was already sent")
	}

	txn := func(tctx context.Context) error {
		token, plain, err := srv.VerificationTokenManager.Create(ctx, user.Id)
		if err != nil {
			return fmt.Errorf("'VerificationTokenManager.Create' failed: %w", err)
		}

		event := domain.NewUserVerification(user.Id, plain, token.ExpiresAt)
		if err := srv.EventPublisher.Publish(tctx, event); err != nil {
			return fmt.Errorf("'EventPublisher.Publish' failed: %w", err)
		}

		return nil
	}
	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("transaction failed: %w", err)
	}

	return ports.SendVerificationResponse{
		Message: "Email was sent successfully",
	}, nil
}

func (srv *service) ForgotPassword(ctx context.Context, req ports.ForgotPasswordRequest) (ports.ForgotPasswordResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("'UsersRepo.GetByEmail' failed: %w", err)
	}

	ok, err := srv.PasswordResetTokenManager.IsLastTokenStillValid(ctx, user.Id)
	if err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("'PasswordResetTokenManager.IsLastTokenStillValid' failed: %w", err)
	}

	if ok {
		return ports.ForgotPasswordResponse{}, errs.NewFailedPrecondition("email was already sent")
	}

	txn := func(tctx context.Context) error {
		token, plain, err := srv.PasswordResetTokenManager.Create(ctx, user.Id)
		if err != nil {
			return fmt.Errorf("'PasswordResetTokenManager.Create' failed: %w", err)
		}

		event := domain.NewUserPasswordReset(user.Id, plain, token.ExpiresAt)
		if err := srv.EventPublisher.Publish(tctx, event); err != nil {
			return fmt.Errorf("'EventPublisher.Publish' failed: %w", err)
		}

		return nil
	}
	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("transaction failed: %w", err)
	}

	return ports.ForgotPasswordResponse{
		Message: "email was send successfully",
	}, nil
}

func (srv *service) ResetPassword(ctx context.Context, req ports.ResetPasswordRequest) (ports.ResetPasswordResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	ok, token, err := srv.PasswordResetTokenManager.Check(ctx, req.Token)
	if err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("'PasswordResetTokenManager.Check' failed: %w", err)
	}

	if !ok {
		return ports.ResetPasswordResponse{}, errs.NewFailedPrecondition("invalid or expired token")
	}

	passwordHash, err := srv.PasswordManager.Hash(req.Password)
	if err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("'PasswordHasher.Hash' failed: %w", err)
	}

	txn := func(tctx context.Context) error {
		user, err := srv.UsersRepo.GetById(tctx, token.UserId)
		if err != nil {
			return fmt.Errorf("'UsersRepo.GetById' failed: %w", err)
		}

		now := time.Now().UTC()
		user.PasswordHash = passwordHash
		user.UpdatedAt = now
		if err := srv.UsersRepo.Update(tctx, user); err != nil {
			return fmt.Errorf("'UsersRepo.Update' failed: %w", err)
		}

		if err := srv.PasswordResetTokenManager.Invalidate(ctx, token); err != nil {
			return fmt.Errorf("'PasswordResetTokenManager.Invalidate' failed: %w", err)
		}

		return nil
	}

	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("transaction failed: %w", err)
	}

	return ports.ResetPasswordResponse{
		Message: "Password was successfully reset",
	}, nil
}

func (srv *service) ChangePassword(ctx context.Context, req ports.ChangePasswordRequest) (ports.ChangePasswordResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.ChangePasswordResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.GetById(ctx, req.UserId)
	if err != nil {
		return ports.ChangePasswordResponse{}, fmt.Errorf("'UsersRepo.GetById' failed: %w", err)
	}

	ok, err := srv.PasswordManager.Verify(req.CurrentPassword, user.PasswordHash)
	if err != nil {
		return ports.ChangePasswordResponse{}, fmt.Errorf("'PasswordManager.Verify' failed: %w", err)
	}
	if !ok {
		return ports.ChangePasswordResponse{}, errs.NewFailedPrecondition("current password isn't correct")
	}

	hash, err := srv.PasswordManager.Hash(req.NewPassword)
	if err != nil {
		return ports.ChangePasswordResponse{}, fmt.Errorf("'PasswordManager.Hash' failed: %w", err)
	}

	user.PasswordHash = hash
	user.UpdatedAt = time.Now().UTC()
	if err := srv.UsersRepo.Update(ctx, user); err != nil {
		return ports.ChangePasswordResponse{}, fmt.Errorf("'UsersRepo.Update' failed: %w", err)
	}

	return ports.ChangePasswordResponse{
		Message: "Password was changed successfully",
	}, nil
}
