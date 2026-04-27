package userssrv

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"golang.org/x/crypto/argon2"
)

type Config struct {
	UsersRepo ports.UsersRepository
}

type service struct {
	Config
}

func New(cfg Config) ports.UsersService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) SignUp(ctx context.Context, req ports.SignUpRequest) (ports.SignUpResponse, error) {
	hash, err := hashPassward(req.Password)
	if err != nil {
		return ports.SignUpResponse{}, fmt.Errorf("'hashPassward' failed: %w", err)
	}
	user := domain.User{
		CreatedAt: time.Now(),
		Username:  req.Username,
		Password:  hash,
	}
	if err := srv.UsersRepo.Create(ctx, &user); err != nil {
		return ports.SignUpResponse{}, fmt.Errorf("'UsersRepo.Create' failed: %w", err)
	}

	return ports.SignUpResponse{
		User:         user,
		FrontendUser: user.Frontend(),
		Message:      "user was created successfully",
	}, nil
}

func (srv *service) SignIn(ctx context.Context, req ports.SignInRequest) (ports.SignInResponse, error) {
	user, err := srv.UsersRepo.GetByName(ctx, req.Username)
	if err != nil {
		return ports.SignInResponse{}, fmt.Errorf("'UsersRepo.GetByName' failed: %w", err)
	}
	ok, err := verifyPassword(req.Password, user.Password)
	if err != nil {
		return ports.SignInResponse{}, fmt.Errorf("'verifyPassword' failed: %w", err)
	}
	if !ok {
		return ports.SignInResponse{}, fmt.Errorf("'verifyPassword' failed: %w", err)
	}
	return ports.SignInResponse{
		User:    user,
		Message: "sign in was successful",
	}, nil
}

// Recommended parameters (RFC 9106)
const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

func hashPassward(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads, b64Salt, b64Hash)
	return encoded, nil
}

func verifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	var memory, time uint32
	var threads uint8
	fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	salt, _ := base64.RawStdEncoding.DecodeString(parts[4])
	decodedHash, _ := base64.RawStdEncoding.DecodeString(parts[5])
	comparisonHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(decodedHash)))
	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}
