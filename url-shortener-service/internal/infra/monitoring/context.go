package monitoring

import (
	"context"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/transportmeta"
	"go.opentelemetry.io/otel/trace"
)

// SpanMetadata contains distributed tracing and correlation information
type SpanMetadata struct {
	CorrelationID string // Correlation ID for request tracking across services
	TraceID       string // OpenTelemetry trace ID in hexadecimal format
	SpanID        string // OpenTelemetry span ID in hexadecimal format
}

// NewContextFromSpanMetadata creates a new context with a remote span context from metadata
func NewContextFromSpanMetadata(base context.Context, meta SpanMetadata) (context.Context, error) {
	traceID, err := trace.TraceIDFromHex(meta.TraceID)
	if err != nil {
		return nil, err
	}

	spanID, err := trace.SpanIDFromHex(meta.SpanID)
	if err != nil {
		return nil, err
	}

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})

	ctx := trace.ContextWithRemoteSpanContext(base, sc)

	if meta.CorrelationID != "" {
		ctx = transportmeta.WithValue(ctx, "correlation_id", meta.CorrelationID)
	}

	return ctx, nil
}

// EnrichContextWithSpanMetadata reconstructs a remote parent span context
// using trace_id, span_id, correlation_id extracted from Kafka payload.
func EnrichContextWithSpanMetadata(ctx context.Context, meta SpanMetadata) (context.Context, error) {
	// Attach correlation ID into ctx first
	if meta.CorrelationID != "" {
		ctx = transportmeta.WithValue(ctx, "correlation_id", meta.CorrelationID)
	}

	// Parse trace_id from hex
	tid, err := trace.TraceIDFromHex(meta.TraceID)
	if err != nil {
		// malformed trace → skip to avoid breaking consumer
		return nil, err
	}

	sid, err := trace.SpanIDFromHex(meta.SpanID)
	if err != nil {
		return nil, err
	}

	// Construct remote parent span context
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})

	// Attach span context to new ctx
	return trace.ContextWithSpanContext(ctx, sc), nil
}
