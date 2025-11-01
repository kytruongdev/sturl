package logging

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey string

const loggerKey ctxKey = "app_logger"

// ToContext keeps the base logger (no span fields baked in)
func ToContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext returns a logger enriched with the CURRENT span from ctx.
// This ensures nested spans show correct span_id in logs, even if the base
// logger was created earlier at middleware time.
func FromContext(ctx context.Context) *zerolog.Logger {
	var base *zerolog.Logger
	if ctx == nil {
		return &nop
	}
	if l, ok := ctx.Value(loggerKey).(Logger); ok && l.Z() != nil {
		base = l.Z()
	} else {
		return &nop
	}

	sc := trace.SpanContextFromContext(ctx)
	// if no active span, just return base
	if !sc.IsValid() {
		return base
	}

	// enrich dynamically per call
	ll := base.With().
		Str("trace_id", sc.TraceID().String()).
		Str("span_id", sc.SpanID().String()).
		Logger()
	return &ll
}
