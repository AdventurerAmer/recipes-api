package infrastructure

import (
	"fmt"
	"time"
)

type Option func(cfg *Config) error

type Config struct {
	startupTimeout  time.Duration
	shutdownTimeout time.Duration
}

func WithStartupTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout == 0 {
			return fmt.Errorf("option: startupTimeout can't be zero")
		}
		cfg.startupTimeout = timeout
		return nil
	}
}

func WithShutdownTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout == 0 {
			return fmt.Errorf("option: shutdownTimeout can't be zero")
		}
		cfg.shutdownTimeout = timeout
		return nil
	}
}
