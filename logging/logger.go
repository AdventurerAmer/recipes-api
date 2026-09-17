package logging

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ThreeDotsLabs/humanslog"
)

type Logger = slog.Logger

func New(opts ...Option) *Logger {
	cfg := Config{
		LocalEnv:  false,
		AddSource: false,
		Level:     LevelDebug,
		Format:    "json",
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	level := slog.LevelDebug
	switch cfg.Level {
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	}

	replaceAttr := replaceAttrNonLocal
	if cfg.LocalEnv {
		replaceAttr = replaceAttrLocal
	}

	handlerOpts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   cfg.AddSource,
		ReplaceAttr: replaceAttr,
	}

	var handler slog.Handler
	if cfg.Format == FormatText {
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
