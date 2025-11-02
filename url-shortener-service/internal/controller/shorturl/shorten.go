package shorturl

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/tracing"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
)

// ShortenInput holds input params for creating short validator
type ShortenInput struct {
	OriginalURL string
}

const (
	// MaxSlugLength defines max length of slug
	MaxSlugLength = 7
)

var generateShortCodeFunc = generateShortCode

// Shorten creates short url
func (i impl) Shorten(ctx context.Context, inp ShortenInput) (model.ShortUrl, error) {
	var err error
	ctx, span := tracing.StartWithName(ctx, "Controller.Shorten")
	defer tracing.End(&span, &err)

	l := logging.FromContext(ctx)
	defer logging.TimeTrack(l, time.Now(), "controller.Shorten")

	shortUrl, err := i.shortUrlRepo.GetByOriginalURL(ctx, inp.OriginalURL)
	if err != nil {
		if errors.Is(err, shorturl.ErrNotFound) {
			l.Warn().Msg("[Shorten] shorten URL not found, starting to create")
			m, err := i.shortUrlRepo.Insert(ctx, model.ShortUrl{
				OriginalURL: inp.OriginalURL,
				Status:      model.ShortUrlStatusActive,
				ShortCode:   generateShortCodeFunc(MaxSlugLength),
			})
			if err != nil {
				l.Error().Err(err).Msg("[Shorten] shortUrlRepo.Insert err")
				return model.ShortUrl{}, err
			}

			l.Info().Msg("[Shorten] shorten URL created")

			return m, nil
		}

		l.Error().Stack().Err(err).Msg("[Shorten] shortUrlRepo.GetByOriginalURL err")

		return model.ShortUrl{}, err
	}

	return shortUrl, nil
}

func generateShortCode(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}

	return string(b)
}
