package monitoring

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Start begins a new span using the global tracer.
// Use defer End(span, &err) after calling it.
func Start(ctx context.Context, name string) (context.Context, trace.Span) {
	return globalTracer.Start(ctx, name)
}

// End finishes a span and records any error if provided.
func End(span trace.Span, errPtr *error) {
	if span == nil {
		return
	}
	if errPtr != nil && *errPtr != nil {
		span.RecordError(*errPtr)
		span.SetStatus(codes.Error, (*errPtr).Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}

// extractTraceInfo returns trace and span IDs from context (if any).
func extractTraceInfo(ctx context.Context) (traceID, spanID string) {
	sc := trace.SpanContextFromContext(ctx)
	if sc.HasTraceID() {
		traceID = sc.TraceID().String()
	}
	if sc.HasSpanID() {
		spanID = sc.SpanID().String()
	}
	return
}
