package monitoring

import (
	"context"
	"errors"

	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring/tracing"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey int

const monitorKey ctxKey = 1

// Monitor provides helper methods for logging and tracing within a request lifecycle
type Monitor struct {
	base logging.Logger
}

// LoggerWithSpan returns a structured logger enriched with trace_id and span_id
// derived from the provided context.
// ctx: used to extract the current span information and is not modified. If the ctx
// does not contain an active span, "unknown" values are used to keep log structure consistent
func (m *Monitor) LoggerWithSpan(ctx context.Context) logging.Logger {
	traceID := "unknown"
	spanID := "unknown"

	sc := trace.SpanContextFromContext(ctx)
	if sc.HasTraceID() {
		traceID = sc.TraceID().String()
		spanID = sc.SpanID().String()
	}

	return m.base.With(
		logging.String("trace_id", traceID),
		logging.String("span_id", spanID),
	)
}

// LoggerWithNonSpan returns a logger detached from tracing context.
// Typically used outside of request scope (e.g., during startup or background jobs).
func (m *Monitor) LoggerWithNonSpan() logging.Logger {
	return m.base
}

// Span retrieves the current active span from the provided context
func (m *Monitor) Span(ctx context.Context) trace.Span { return trace.SpanFromContext(ctx) }

// StartSpan begins a new tracing span with the specified name and returns
// the updated context along with the span
func (m *Monitor) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return tracing.Start(ctx, name)
}

// StartSpanWithLog begins a new tracing span with the given name and returns
// the updated context, the span, and a logger enriched with the span's trace_id
// and span_id. This is the preferred way to start new traced operations.
func (m *Monitor) StartSpanWithLog(ctx context.Context, name string) (context.Context, trace.Span, logging.Logger) {
	ctx, span := tracing.Start(ctx, name)
	l := m.LoggerWithSpan(ctx)
	return ctx, span, l
}

// EndSpan safely ends the provided span and records any error found in errPtr.
// This should typically be deferred right after starting a span.
func (m *Monitor) EndSpan(span *trace.Span, errPtr *error) { tracing.End(span, errPtr) }

// ToContext attaches a Monitor instance to the given context, allowing
// it to be retrieved later using FromContext
func ToContext(ctx context.Context, m *Monitor) context.Context {
	return context.WithValue(ctx, monitorKey, m)
}

// FromContext retrieves a Monitor from the provided context
func FromContext(ctx context.Context) *Monitor {
	if v := ctx.Value(monitorKey); v != nil {
		if m, ok := v.(*Monitor); ok {
			return m
		}
	}
	l := logging.FromContext(ctx)
	return &Monitor{base: l}
}

// MergeError combines two error values into a single error using errors.Join
func MergeError(dst *error, err error) {
	if err == nil {
		return
	}
	if *dst == nil {
		*dst = err
		return
	}
	*dst = errors.Join(*dst, err)
}
