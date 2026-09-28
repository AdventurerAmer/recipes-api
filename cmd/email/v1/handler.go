package v1

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/AdventurerAmer/recipes-api/mailer"
	"github.com/AdventurerAmer/recipes-api/validation"
)

type eventHandler struct {
	usersRepo ports.UsersRepository
	templates *template.Template
	mailer    *mailer.Mailer
}

func newEventHandler(usersRepo ports.UsersRepository, templates *template.Template, mailer *mailer.Mailer) *eventHandler {
	return &eventHandler{
		usersRepo: usersRepo,
		templates: templates,
		mailer:    mailer,
	}
}

func (h *eventHandler) OnUserCreated(ctx context.Context, e domain.UserCreatedEvent) error {
	if err := validation.Validate(e); err != nil {
		return err
	}
	return h.verifyUser(ctx, e.UserId)
}

func (h *eventHandler) OnUserVerification(ctx context.Context, e domain.UserVerificationEvent) error {
	if err := validation.Validate(e); err != nil {
		return err
	}
	return h.verifyUser(ctx, e.UserId)
}

func (h *eventHandler) OnUserPasswordReset(ctx context.Context, e domain.UserPasswordResetEvent) error {
	if err := validation.Validate(e); err != nil {
		return err
	}

	user, err := h.usersRepo.GetById(ctx, e.UserId)
	if err != nil {
		return fmt.Errorf("'usersRepo.GetById' failed: %w", err)
	}

	if user.PasswordHash != "" {
		return nil
	}

	expiresAt := user.ForgotPassword.ExpiresAt
	now := time.Now().UTC()
	if now.After(expiresAt) {
		return nil
	}

	data := struct {
		Name      string
		ResetURL  string
		ExpiresIn time.Duration
	}{
		Name:      user.DisplayName,
		ResetURL:  fmt.Sprintf("http://localhost:3000/api/users/reset-password?token=%s", user.ForgotPassword.Token),
		ExpiresIn: expiresAt.Sub(now),
	}

	// TODO: have a buffer bool here...
	var b bytes.Buffer
	if err := h.templates.ExecuteTemplate(&b, "reset.html", data); err != nil {
		return fmt.Errorf("'templates.ExecuteTemplate' failed: %w", err)
	}

	message := mailer.Message{
		To:      user.Email,
		Subject: "Reset your password",
		Body:    b.String(),
	}
	if err := h.mailer.Send(message); err != nil {
		return fmt.Errorf("'mailer.Send' failed: %w", err)
	}

	return nil
}

func (h *eventHandler) verifyUser(ctx context.Context, userId string) error {
	user, err := h.usersRepo.GetById(ctx, userId)
	if err != nil {
		return fmt.Errorf("'usersRepo.GetById' failed: %w", err)
	}

	if user.IsVerified {
		return nil
	}

	expiresAt := user.Verification.ExpiresAt
	now := time.Now().UTC()
	if now.After(expiresAt) {
		return nil
	}

	data := struct {
		Name      string
		VerifyURL string
		ExpiresIn time.Duration
	}{
		Name:      user.DisplayName,
		VerifyURL: fmt.Sprintf("http://localhost:3000/api/users/verify?token=%s", user.Verification.Token),
		ExpiresIn: expiresAt.Sub(now),
	}

	// TODO: have a buffer bool here...
	var b bytes.Buffer
	if err := h.templates.ExecuteTemplate(&b, "verify.html", data); err != nil {
		return fmt.Errorf("'templates.ExecuteTemplate' failed: %w", err)
	}

	message := mailer.Message{
		To:      user.Email,
		Subject: "Verify your account",
		Body:    b.String(),
	}
	if err := h.mailer.Send(message); err != nil {
		return fmt.Errorf("'mailer.Send' failed: %w", err)
	}

	return nil
}
