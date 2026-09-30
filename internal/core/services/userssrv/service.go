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
	VerificationTokenExpiresAfter   time.Duration
	ForgotpasswordTokenExpiresAfter time.Duration
	PasswordManager                 ports.PasswordManager
	Transactor                      ports.Transactor
	UsersRepo                       ports.UsersRepository
	TokensRepo                      ports.TokensRepository
	EventPublisher                  ports.EventPublisher
}

type service struct {
	Config
}

func New(cfg Config) ports.UsersService {
	// TODO: add this to config
	if cfg.VerificationTokenExpiresAfter == 0 {
		cfg.VerificationTokenExpiresAfter = 10 * time.Minute
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
		tokenStr, tokenHash, err := tokens.Generate()
		if err != nil {
			return fmt.Errorf("'tokens.Generate' failed: %w", err)
		}

		if err := srv.UsersRepo.Create(tctx, user); err != nil {
			return fmt.Errorf("'UsersRepo.Create' failed: %w", err)
		}

		expiresAt := now.Add(srv.VerificationTokenExpiresAfter)
		token := &domain.Token{
			CreatedAt: now,
			Type:      domain.TokenTypeVerification,
			UserId:    user.Id,
			ExpiresAt: expiresAt,
			Hash:      tokenHash,
		}
		if err := srv.TokensRepo.Create(tctx, token); err != nil {
			return fmt.Errorf("'TokensRepo.Create' failed: %w", err)
		}

		event := domain.NewUserCreated(user.Id, tokenStr, expiresAt)
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

func (srv *service) Verify(ctx context.Context, req ports.VerifyUserRequest) (ports.VerifyResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.VerifyResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	hash := tokens.Hash(req.Token)
	token, err := srv.TokensRepo.Get(ctx, domain.TokenTypeVerification, hash)
	if err != nil {
		if errs.IsNotFound(err) {
			return ports.VerifyResponse{}, errs.NewFailedPrecondition("invalid or expired token")
		}
		return ports.VerifyResponse{}, fmt.Errorf("'TokensRepo.Get' failed: %w", err)
	}

	now := time.Now().UTC()
	if now.After(token.ExpiresAt) {
		return ports.VerifyResponse{}, errs.NewFailedPrecondition("invalid or expired token")
	}

	txn := func(tctx context.Context) error {
		user, err := srv.UsersRepo.GetById(tctx, token.UserId)
		if err != nil {
			return fmt.Errorf("'UsersRepo.GetById' failed: %w", err)
		}

		user.IsVerified = true
		user.UpdatedAt = now
		if err := srv.UsersRepo.Update(tctx, user); err != nil {
			return fmt.Errorf("'UsersRepo.Update' failed: %w", err)
		}

		token.UsedAt = &now
		if err := srv.TokensRepo.Update(tctx, token); err != nil {
			return fmt.Errorf("'TokensRepo.Update' failed: %w", err)
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

	lastToken, err := srv.TokensRepo.GetLast(ctx, user.Id, domain.TokenTypeVerification)
	if err != nil {
		if !errs.IsNotFound(err) {
			return ports.SendVerificationResponse{}, fmt.Errorf("'TokensRepo.GetLast' failed: %w", err)
		}
	}

	now := time.Now().UTC()
	if lastToken != nil && lastToken.ExpiresAt.After(now) {
		return ports.SendVerificationResponse{}, errs.NewFailedPrecondition("email was already sent")
	}

	tokenStr, hash, err := tokens.Generate()
	if err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("'tokens.Generate' failed: %w", err)
	}

	expiresAt := now.Add(srv.VerificationTokenExpiresAfter)

	txn := func(tctx context.Context) error {
		token := domain.Token{
			CreatedAt: now,
			Type:      domain.TokenTypeVerification,
			Hash:      hash,
			ExpiresAt: expiresAt,
			UserId:    user.Id,
		}

		if err := srv.TokensRepo.Create(tctx, &token); err != nil {
			return fmt.Errorf("'TokensRepo.Create' failed: %w", err)
		}

		event := domain.NewUserVerification(user.Id, tokenStr, expiresAt)
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

	lastToken, err := srv.TokensRepo.GetLast(ctx, user.Id, domain.TokenTypePasswordReset)
	if err != nil {
		if !errs.IsNotFound(err) {
			return ports.ForgotPasswordResponse{}, fmt.Errorf("'TokensRepo.GetLast' failed: %w", err)
		}
	}

	now := time.Now().UTC()
	if lastToken != nil && lastToken.ExpiresAt.After(now) {
		return ports.ForgotPasswordResponse{}, errs.NewFailedPrecondition("email was already sent")
	}

	tokenStr, hash, err := tokens.Generate()
	if err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("'tokens.Generate' failed: %w", err)
	}

	expiresAt := now.Add(srv.ForgotpasswordTokenExpiresAfter)

	txn := func(tctx context.Context) error {
		token := domain.Token{
			CreatedAt: now,
			Type:      domain.TokenTypeVerification,
			Hash:      hash,
			ExpiresAt: expiresAt,
			UserId:    user.Id,
		}

		if err := srv.TokensRepo.Create(tctx, &token); err != nil {
			return fmt.Errorf("'UsersRepo.Update' failed: %w", err)
		}

		event := domain.NewUserPasswordReset(user.Id, tokenStr, expiresAt)
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

	tokenHash := tokens.Hash(req.Token)
	token, err := srv.TokensRepo.Get(ctx, domain.TokenTypePasswordReset, tokenHash)
	if err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("'TokensRepo.Get: %w", err)
	}

	now := time.Now().UTC()
	if now.After(token.ExpiresAt) {
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

		user.PasswordHash = passwordHash
		user.UpdatedAt = now
		if err := srv.UsersRepo.Update(tctx, user); err != nil {
			return fmt.Errorf("'UsersRepo.Update' failed: %w", err)
		}

		token.UsedAt = &now
		if err := srv.TokensRepo.Update(tctx, token); err != nil {
			return fmt.Errorf("'TokensRepo.Update' failed: %w", err)
		}

		return nil
	}

	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("transaction failed: %w", err)
	}

	return ports.ResetPasswordResponse{}, nil
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
