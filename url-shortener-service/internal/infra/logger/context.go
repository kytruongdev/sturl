package logger

import (
	"context"

	"github.com/rs/zerolog"
)

type ctxKey string

const (
	loggerKey        ctxKey = "logger"
	correlationIDKey ctxKey = "correlation_id"
)

// WithLogger embeds a request-scoped logger into context.
func WithLogger(ctx context.Context, l *zerolog.Logger) context.Context {
	if ctx == nil || l == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext fetches the logger stored in context (if any).
func FromContext(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		return nil
	}
	if l, ok := ctx.Value(loggerKey).(*zerolog.Logger); ok {
		return l
	}

	return nil
}

// WithCorrelationID stores correlation/request id into context.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	if ctx == nil || id == "" {
		return ctx
	}
	return context.WithValue(ctx, correlationIDKey, id)
}

// CorrelationIDFromContext returns correlation_id if present.
func CorrelationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

// UpdateLogger updates the logger inside a context by applying fn to its context.
// Safe to call with nil logger (no-op).
func UpdateLogger(ctx context.Context, fn func(zerolog.Context) zerolog.Context) {
	if fn == nil {
		return
	}
	if l := FromContext(ctx); l != nil {
		l.UpdateContext(fn)
	}
}
