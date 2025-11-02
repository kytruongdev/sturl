package shorturl

import (
	"context"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/tracing"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
)

// Retrieve retrieves short url by short code
func (i impl) Retrieve(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	var err error
	ctx, span := tracing.StartWithName(ctx, "Controller.Retrieve")
	defer tracing.End(&span, &err)

	l := logging.FromContext(ctx)
	defer logging.TimeTrack(l, time.Now(), "controller.Retrieve")

	m, err := i.shortUrlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		l.Error().Err(err).Msg("[Retrieve] shortUrlRepo.GetByShortCode err")
		return model.ShortUrl{}, err
	}

	if m.Status != model.ShortUrlStatusActive {
		return model.ShortUrl{}, ErrInactiveURL
	}

	return m, err
}
