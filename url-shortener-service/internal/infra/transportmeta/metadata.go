package transportmeta

import (
	"context"
	"net/http"
)

// Metadata represents request-scoped identifiers used for tracing and logging.
// It carries request-scoped identifiers that are typically propagated through HTTP headers.
type Metadata struct {
	CorrelationID string
	RequestID     string
	Extra         map[string]string
}

type ctxKey struct{}

// ToContext stores the provided Metadata in the given context, allowing it
// to be retrieved later by FromContext. It replaces any existing metadata
func ToContext(ctx context.Context, m Metadata) context.Context {
	return context.WithValue(ctx, ctxKey{}, m)
}

// FromContext retrieves Metadata from the given context.
// If no metadata is present, it returns an empty Metadata struct.
func FromContext(ctx context.Context) Metadata {
	if v := ctx.Value(ctxKey{}); v != nil {
		if m, ok := v.(Metadata); ok {
			return m
		}
	}
	return Metadata{}
}

// ExtractFromRequest builds a Metadata struct from the standard correlation
// and request ID headers in the given HTTP request
func ExtractFromRequest(r *http.Request) Metadata {
	return Metadata{
		CorrelationID: r.Header.Get("X-Correlation-ID"),
		RequestID:     r.Header.Get("X-Request-ID"),
	}
}

// getByKey returns a metadata value by its header key.
// If the key does not match a known field, Extra is checked.
func (m *Metadata) getByKey(key string) string {
	switch key {
	case "X-Correlation-ID":
		return m.CorrelationID
	case "X-Request-ID":
		return m.RequestID
	default:
		return m.Extra[key]
	}
}

// setByKey sets a metadata value by its header key.
// If the key does not match a known field, it is stored in Extra.
func (m *Metadata) setByKey(key, val string) {
	if m.Extra == nil {
		m.Extra = map[string]string{}
	}
	switch key {
	case "X-Correlation-ID":
		m.CorrelationID = val
	case "X-Request-ID":
		m.RequestID = val
	default:
		m.Extra[key] = val
	}
}
