package logger

import (
	"context"

	"github.com/rs/zerolog"
)

type ctxKey string

const loggerKey ctxKey = "app_logger"

// ToContext embeds the logger into a context
func ToContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext retrieves *zerolog.Logger directly (unwraps .Z())
// This makes it easy to log directly: logger.FromContext(ctx).Info().Msg("...")
func FromContext(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		return &nop
	}
	if l, ok := ctx.Value(loggerKey).(Logger); ok {
		return l.Z()
	}
	return &nop
}
