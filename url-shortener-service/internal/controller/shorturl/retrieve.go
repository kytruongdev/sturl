package shorturl

import (
	"context"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
)

// Retrieve retrieves the original URL associated with the given short code.
func (i impl) Retrieve(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	var err error
	ctx, span := monitoring.Start(ctx, "Controller.Retrieve")
	defer monitoring.End(span, &err)

	l := monitoring.Log(ctx)

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
