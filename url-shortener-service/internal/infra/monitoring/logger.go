package monitoring

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/transportmeta"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

// Field represents a function type for enriching structured logs with typed key-value pairs.
type Field func(zerolog.Context) zerolog.Context

// Logger wraps zerolog for structured logging.
type Logger struct {
	z   zerolog.Logger
	ctx zerolog.Context
}

// NewLogger constructs the base zerolog logger.
func NewLogger(cfg Config) Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, NoColor: !cfg.LogPretty}
	builder := zerolog.New(output).With().Timestamp().
		Str("service", cfg.ServiceName).
		Str("env", cfg.Env)
	return Logger{z: builder.Logger(), ctx: builder}
}

// Log returns a structured logger enriched with trace and correlation data.
func Log(ctx context.Context) Logger {
	traceID, spanID := extractTraceInfo(ctx)
	meta := transportmeta.FromContext(ctx)
	var reqID string
	if reqID = meta.RequestID; reqID == "" {
		reqID = xid.New().String()
	}

	builder := globalLogger.With()

	if traceID != "" {
		builder = builder.Str("trace_id", traceID)
	}
	if spanID != "" {
		builder = builder.Str("span_id", spanID)
	}
	if meta.CorrelationID != "" {
		builder = builder.Str("correlation_id", meta.CorrelationID)
	}
	if reqID != "" {
		builder = builder.Str("request_id", reqID)
	}

	return Logger{z: builder.Logger(), ctx: builder}
}

// Info returns a zerolog event for logging at INFO level.
func (l Logger) Info() *zerolog.Event { return l.z.Info() }

// Error returns a zerolog event for logging at ERROR level.
func (l Logger) Error() *zerolog.Event { return l.z.Error() }

// Warn returns a zerolog event for logging at WARN level.
func (l Logger) Warn() *zerolog.Event { return l.z.Warn() }

// Debug returns a zerolog event for logging at DEBUG level.
func (l Logger) Debug() *zerolog.Event { return l.z.Debug() }

// With creates a new Logger derived from the current one and enriched
// with the provided structured fields. Each Field function appends
// a typed key-value pair to the log context.
func (l Logger) With(fields ...Field) Logger {
	b := l.z.With()
	for _, f := range fields {
		b = f(b)
	}

	return Logger{z: b.Logger()}
}

// TimeTrack logs the duration elapsed since the provided start time
func (l Logger) TimeTrack(start time.Time, label string, fields ...Field) {
	elapsed := time.Since(start)
	l.With(fields...).Info().Dur("elapsed", elapsed).Msg(label)
}

func (l Logger) Fields(fields map[string]interface{}) Logger {
	b := l.ctx

	for key, value := range fields {
		switch v := value.(type) {
		case string:
			b = b.Str(key, v)
		case fmt.Stringer:
			b = b.Str(key, v.String())
		case int:
			b = b.Int(key, v)
		case int64:
			b = b.Int64(key, v)
		case uint:
			b = b.Uint(key, v)
		case uint64:
			b = b.Uint64(key, v)
		case float64:
			b = b.Float64(key, v)
		case bool:
			b = b.Bool(key, v)
		case time.Time:
			b = b.Time(key, v)
		case []string:
			b = b.Strs(key, v)
		default:
			b = b.Interface(key, v)
		}
	}

	return Logger{
		z:   b.Logger(),
		ctx: b,
	}
}

func (l Logger) Field(key string, value interface{}) Logger {
	b := l.ctx

	switch v := value.(type) {
	case string:
		b = b.Str(key, v)
	case fmt.Stringer:
		b = b.Str(key, v.String())
	case int:
		b = b.Int(key, v)
	case int64:
		b = b.Int64(key, v)
	case uint:
		b = b.Uint(key, v)
	case uint64:
		b = b.Uint64(key, v)
	case float64:
		b = b.Float64(key, v)
	case bool:
		b = b.Bool(key, v)
	case time.Time:
		b = b.Time(key, v)
	case []string:
		b = b.Strs(key, v)
	default:
		b = b.Interface(key, v)
	}

	return Logger{
		z:   b.Logger(),
		ctx: b,
	}
}
