package common

const (
	// HeaderCorrelationID defines the standard HTTP header key used
	// to correlate logs and traces across multiple services in the same request flow.
	HeaderCorrelationID = "X-Correlation-ID"

	// HeaderRequestID defines the standard HTTP header key used
	// to uniquely identify each individual HTTP request within a service.
	HeaderRequestID = "X-Request-ID"

	// OpInbound represents the operation name used by the OpenTelemetry
	// tracing middleware for inbound HTTP requests handled by the Url Shortener service.
	// It identifies the root span created when a request first enters this service.
	OpInbound = "url-shortener.inbound"
)
