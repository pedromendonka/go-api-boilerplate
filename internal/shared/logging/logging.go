// Package logging provides structured logging using slog.
package logging

import (
	"context"
	"log/slog"
	"os"
)

// contextKey is used to store logger in context.
type contextKey struct{}

// Config holds logging configuration.
type Config struct {
	Level  slog.Level // Log level (Debug, Info, Warn, Error)
	Format string     // "json" or "text"
}

// DefaultConfig returns sensible defaults for production.
func DefaultConfig() Config {
	return Config{
		Level:  slog.LevelInfo,
		Format: "json",
	}
}

// New creates a configured slog.Logger.
func New(cfg Config) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	switch cfg.Format {
	case "text":
		// Use colored handler for terminal output
		handler = NewColoredHandler(os.Stdout, opts)
	default:
		// Use JSON handler for production/log aggregation
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// WithContext returns a new context with the logger attached.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext retrieves the logger from context, or returns the default logger.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(contextKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// WithAttrs returns a logger with additional attributes attached.
func WithAttrs(logger *slog.Logger, attrs ...slog.Attr) *slog.Logger {
	return slog.New(logger.Handler().WithAttrs(attrs))
}
