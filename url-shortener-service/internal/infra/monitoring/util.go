package monitoring

import (
	"context"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/transportmeta"
)

// SpanMetadataFromContext extracts trace, span, and correlation metadata from the context
func SpanMetadataFromContext(ctx context.Context) SpanMetadata {
	traceID, spanID := extractTraceInfo(ctx)
	meta := transportmeta.FromContext(ctx)

	return SpanMetadata{
		CorrelationID: meta.CorrelationID,
		TraceID:       traceID,
		SpanID:        spanID,
	}
}
