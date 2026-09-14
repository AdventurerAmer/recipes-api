package config

import "time"

type Authentication struct {
	Secret       string        `koanf:"secret"`
	MaxIdelConns int           `koanf:"maxIdelConns"`
	Name         string        `koanf:"name"`
	MaxAge       time.Duration `koanf:"maxAge"`
}
