package logging

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

type ctxKey int

const loggerKey ctxKey = 1

// ToContext attaches a Logger to the given context
func ToContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext extracts a Logger from the provided context
func FromContext(ctx context.Context) Logger {
	if v := ctx.Value(loggerKey); v != nil {
		if l, ok := v.(Logger); ok {
			return l
		}
	}
	cw := zerolog.ConsoleWriter{Out: os.Stdout}
	core := zerolog.New(cw).With().Timestamp().Str("service", "unknown").Logger()
	return Logger{z: &core}
}
