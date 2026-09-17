package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/AdventurerAmer/recipes-api/errs"
	"github.com/AdventurerAmer/recipes-api/logging"
	"github.com/AdventurerAmer/recipes-api/validation"
	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Env           Env            `koanf:"env" validate:"oneof=local staging production"`
	App           App            `koanf:"app"`
	Infra         Infra          `koanf:"infra"`
	Auth          Authentication `koanf:"auth"`
	Observability Observability  `koanf:"observability"`
	Services      Services       `koanf:"services"`
	Constants     Constants      `koanf:"constants"`
}

func (cfg *Config) NewLogger() *logging.Logger {
	logger := logging.New(
		logging.WithLocalEnv(cfg.Env == EnvLocal),
		logging.WithFormat(cfg.Observability.Logging.Format),
		logging.WithLevel(cfg.Observability.Logging.Level),
		logging.WithAddSource(*cfg.Observability.Logging.AddSource),
	)
	return logger
}

func Load() (*Config, error) {
	delim := "."
	k := koanf.New(delim)

	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load config.yaml: %w", err)
	}

	var envFile string
	flag.StringVar(&envFile, "env-file", ".env.production", "env file to load config from")
	flag.Parse()

	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	envPrefix := "RECIPES."
	envOpt := env.Opt{
		Prefix: envPrefix,
		TransformFunc: func(k, v string) (string, any) {
			k = strings.TrimPrefix(k, envPrefix)
			k = strings.ToLower(k)
			keyParts := strings.Split(k, delim)
			for idx, part := range keyParts {
				keyParts[idx] = snakeCaseToCamelCase(part)
			}
			k = strings.Join(keyParts, delim)

			if strings.Contains(v, " ") {
				valParts := strings.Split(v, " ")
				if len(valParts) > 1 {
					return k, valParts
				}
			}

			return k, v
		},
	}

	if err := k.Load(env.Provider(delim, envOpt), nil); err != nil {
		return nil, fmt.Errorf("failed to parse env vars: %w", err)
	}

	unmarshalConf := koanf.UnmarshalConf{
		Tag: "koanf",
	}
	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, unmarshalConf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	setDefaults(&cfg)

	if err := validation.Validate(cfg); err != nil {
		if errs.IsValidation(err) {
			var e *errs.Error
			errors.As(err, &e)
			return nil, fmt.Errorf("failed to validate config: %w", fmt.Errorf("one or more invalid fields: %+v", e.Fields))
		}
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Env == "" {
		cfg.Env = EnvLocal
	}

	if cfg.App.Name == "" {
		cfg.App.Name = "Recipes"
	}

	if cfg.App.Version == "" {
		cfg.App.Version = "0.1.0"
	}

	if cfg.Observability.Logging.Level == "" {
		if cfg.Env == EnvLocal {
			cfg.Observability.Logging.Level = "debug"
		} else {
			cfg.Observability.Logging.Level = "info"
		}
	}

	if cfg.Observability.Logging.Format == "" {
		if cfg.Env == EnvLocal {
			cfg.Observability.Logging.Format = "text"
		} else {
			cfg.Observability.Logging.Format = "json"
		}
	}

	if cfg.Observability.Logging.AddSource == nil {
		addSource := (cfg.Env == EnvLocal || cfg.Env == EnvStaging)
		cfg.Observability.Logging.AddSource = &addSource
	}

	if cfg.Observability.HealthChecks.Interval == 0 {
		cfg.Observability.HealthChecks.Interval = 30 * time.Second
	}

	if cfg.Observability.HealthChecks.Timeout == 0 {
		cfg.Observability.HealthChecks.Timeout = 5 * time.Second
	}

	setServiceDefaults(&cfg.Services.Recipes)
}

func snakeCaseToCamelCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")

	builder := &strings.Builder{}
	builder.WriteString(parts[0])

	for _, part := range parts[1:] {
		first, size := utf8.DecodeRuneInString(part)
		builder.WriteRune(unicode.ToUpper(first))
		builder.WriteString(part[size:])
	}

	return builder.String()
}
