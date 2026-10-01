package config

import "time"

type Token struct {
	Secret       string        `koanf:"secret" validate:"required"`
	ExpiresAfter time.Duration `koanf:"expiresAfter" validate:"required,min=1s"`
}

type Tokens struct {
	Verification  Token `koanf:"verification"`
	PasswordReset Token `koanf:"passwordReset"`
}
