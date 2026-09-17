package logging

type Config struct {
	LocalEnv  bool
	AddSource bool
	Level     Level
	Format    Format
}

type Option func(cfg *Config)

func WithLocalEnv(localEnv bool) Option {
	return func(cfg *Config) {
		cfg.LocalEnv = localEnv
	}
}

func WithAddSource(addSource bool) Option {
	return func(cfg *Config) {
		cfg.AddSource = addSource
	}
}

func WithLevel(level Level) Option {
	return func(cfg *Config) {
		cfg.Level = level
	}
}

func WithFormat(format Format) Option {
	return func(cfg *Config) {
		cfg.Format = format
	}
}
