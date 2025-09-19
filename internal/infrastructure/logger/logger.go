package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger interface for structured logging
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	With(args ...any) Logger
	WithContext(ctx context.Context) Logger
}

// SlogLogger implements Logger using slog
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new structured logger
func NewSlogLogger() Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &SlogLogger{
		logger: slog.New(handler),
	}
}

// Info logs an info message
func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message
func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message
func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// Debug logs a debug message
func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// With returns a new logger with additional attributes
func (l *SlogLogger) With(args ...any) Logger {
	return &SlogLogger{
		logger: l.logger.With(args...),
	}
}

// WithContext returns a new logger with context
func (l *SlogLogger) WithContext(ctx context.Context) Logger {
	return &SlogLogger{
		logger: l.logger.With("trace_id", GetTraceIDFromContext(ctx)),
	}
}

// typed key to avoid context key collisions
type traceIDKey struct{}

// SetTraceIDOnContext returns a new context with the given trace id stored safely.
func SetTraceIDOnContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, id)
}

// GetTraceIDFromContext retrieves a trace id from context in a type-safe way.
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value(traceIDKey{}); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return "unknown"
}
