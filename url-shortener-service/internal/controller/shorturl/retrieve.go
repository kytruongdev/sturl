package shorturl

import (
	"context"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"go.opentelemetry.io/otel"
)

// Retrieve retrieves short url by short code
func (i impl) Retrieve(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	tracer := otel.Tracer("url-shortener.controller")
	ctx, span := tracer.Start(ctx, "Controller.Retrieve")
	defer span.End()

	l := logging.FromContext(ctx)
	defer logging.TimeTrack(l, time.Now(), "controller.Retrieve")

	m, err := i.shortUrlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		span.RecordError(err)
		l.Error().Err(err).Msg("[Retrieve] shortUrlRepo.GetByShortCode err")
		return model.ShortUrl{}, err
	}

	if m.Status != model.ShortUrlStatusActive {
		span.RecordError(ErrInactiveURL)
		return model.ShortUrl{}, ErrInactiveURL
	}

	return m, err
}
