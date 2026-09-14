package logging

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ThreeDotsLabs/humanslog"
)

type Level = slog.Level
type Logger = slog.Logger

type Config struct {
	IsLocalEnv bool
	Level      Level
	AddSource  bool
	Format     string
}

func New(cfg *Config) *Logger {
	if cfg == nil {
		cfg = &Config{
			IsLocalEnv: false,
			Level:      slog.LevelError,
			AddSource:  false,
			Format:     "json",
		}
	}

	replaceAttr := replaceAttrNonLocal
	if cfg.IsLocalEnv {
		replaceAttr = replaceAttrLocal
	}

	handlerOpts := &slog.HandlerOptions{
		Level:       cfg.Level,
		AddSource:   cfg.AddSource,
		ReplaceAttr: replaceAttr,
	}

	var handler slog.Handler
	if cfg.Format == "text" {
		opts := &humanslog.Options{
			HandlerOptions:    handlerOpts,
			SortKeys:          true,
			TimeFormat:        "[15:04:05]",
			NewLineAfterLog:   true,
			DebugColor:        humanslog.Magenta,
			StringerFormatter: true,
		}
		handler = humanslog.NewHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

type loggerCtxKey struct{}

func Set(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, logger)
}

func Get(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerCtxKey{}).(*Logger); ok {
		return logger
	}
	return slog.Default()
}

func ParseLevel(level string) Level {
	switch level {
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

func replaceAttrLocal(groups []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case slog.SourceKey:
		if source, ok := attr.Value.Any().(*slog.Source); ok {
			wd, err := os.Getwd()
			if err != nil {
				return attr
			}
			rel, err := filepath.Rel(wd, source.File)
			if err != nil {
				return attr
			}
			source.File = rel
		}
	}
	return attr
}

func replaceAttrNonLocal(groups []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case "password", "secret", "token", "jwt_secret":
		return slog.String(attr.Key, "[REDACTED]")
	}
	return attr
}
