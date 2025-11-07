package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Start begins a new child span from the current context using the given name
func Start(ctx context.Context, name string) (context.Context, trace.Span) {
	tr := otel.Tracer("sturl/monitoring")
	return tr.Start(ctx, name)
}

// End safely ends the span in the given context. If an error is provided, it will be recorded as part of the span's status.
func End(span *trace.Span, errPtr *error) {
	if span == nil || *span == nil {
		return
	}
	if errPtr != nil && *errPtr != nil {
		(*span).RecordError(*errPtr)
		(*span).SetStatus(codes.Error, (*errPtr).Error())
	} else {
		(*span).SetStatus(codes.Ok, "")
	}
	(*span).End()
}
