package shorturl

import (
	"context"
	"errors"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
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

		if errors.Is(err, shorturl.ErrNotFound) {
			return model.ShortUrl{}, ErrURLNotfound
		}

		return model.ShortUrl{}, err
	}

	if m.Status != model.ShortUrlStatusActive {
		return model.ShortUrl{}, ErrInactiveURL
	}

	return m, err
}
