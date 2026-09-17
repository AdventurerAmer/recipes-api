package config

import "time"

type Authentication struct {
	Secret       string        `koanf:"secret" validate:"required,max=128"`
	MaxIdelConns int           `koanf:"maxIdelConns" validate:"required,min=1,max=32"`
	Name         string        `koanf:"name" validate:"required,max=128"`
	MaxAge       time.Duration `koanf:"maxAge" validate:"required,min=1s"`
}
