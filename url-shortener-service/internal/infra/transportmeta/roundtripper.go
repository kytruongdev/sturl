package transportmeta

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// WrapTransport wraps an HTTP RoundTripper to inject metadata and trace context into outbound requests.
// It injects metadata (Request-ID, Correlation-ID) and trace context into outbound HTTP requests.
func WrapTransport(next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return roundTripper{next: next}
}

// roundTripper wraps an http.RoundTripper to inject metadata and trace context into outbound HTTP requests.
type roundTripper struct{ next http.RoundTripper }

// RoundTrip implements the http.RoundTripper interface.
// It injects standard metadata headers (correlation ID, request ID)
// and W3C trace context IDs (trace-id, span-id) into outbound requests
// before delegating to the next transport layer.
//
// Note: The OpenTelemetry instrumentation (otelhttp) will also inject the
// "traceparent" header automatically; these plain IDs are provided mainly
// for log correlation and debugging purposes.
func (r roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	meta := FromContext(req.Context())

	if meta.CorrelationID != "" {
		req.Header.Set("X-Correlation-ID", meta.CorrelationID)
	}
	if meta.RequestID != "" {
		req.Header.Set("X-Request-ID", meta.RequestID)
	}

	// optional: attach W3C trace context ids as plain headers (otelhttp will also inject traceparent)
	if sc := trace.SpanContextFromContext(req.Context()); sc.HasTraceID() {
		req.Header.Set("trace-id", sc.TraceID().String())
		if sc.HasSpanID() {
			req.Header.Set("span-id", sc.SpanID().String())
		}
	}

	return r.next.RoundTrip(req)
}
