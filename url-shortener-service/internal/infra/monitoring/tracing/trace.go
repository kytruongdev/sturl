package tracing

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Start auto-creates a span named after the caller function.
// Example: ctx, span := tracing.Start(ctx); defer tracing.End(&span, &err)
func Start(ctx context.Context) (context.Context, trace.Span) {
	pc, _, _, ok := runtime.Caller(1)
	name := "unknown"
	if ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			name = simplifyName(fn.Name())
		}
	}

	tracer := otel.Tracer("sturl")
	ctx, span := tracer.Start(ctx, name)
	return ctx, span
}

// StartWithName starts a span with a custom name instead of auto-detected one.
func StartWithName(ctx context.Context, name string) (context.Context, trace.Span) {
	tracer := otel.Tracer("sturl")
	return tracer.Start(ctx, name)
}

// End ends the span, recording error status if err != nil.
// Use with defer: defer tracing.End(&span, &err)
func End(span *trace.Span, err *error) {
	if *err != nil {
		(*span).RecordError(*err)
		(*span).SetStatus(codes.Error, fmt.Sprintf("%v", *err))
	}
	(*span).End()
}

// simplifyName trims long package paths to show only the function name
func simplifyName(full string) string {
	if full == "" {
		return "unknown"
	}
	// Remove everything before last '/'
	if i := strings.LastIndex(full, "/"); i != -1 {
		full = full[i+1:]
	}
	// Remove method receiver prefix like "(*Handler).Shorten"
	if i := strings.Index(full, ")."); i != -1 {
		full = full[i+2:]
	}
	return full
}
