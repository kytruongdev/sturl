package monitoring

import (
	"context"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/transportmeta"
	"github.com/rs/xid"
)

func NewContext() context.Context {
	// Generate new per-batch IDs
	reqID := xid.New().String()

	md := transportmeta.Metadata{
		RequestID:     reqID,
		CorrelationID: "", // empty is correct for producer
		Extra:         map[string]string{},
	}

	return transportmeta.ToContext(context.Background(), md)
}
