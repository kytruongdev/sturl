package shorturl

import (
	"context"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/logger"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
)

// Retrieve retrieves short url by short code
func (i impl) Retrieve(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	l := logger.FromContext(ctx)

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
