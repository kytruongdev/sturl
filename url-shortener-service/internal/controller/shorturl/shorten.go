package shorturl

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/id"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
	pkgerrors "github.com/pkg/errors"
)

// ShortenInput represents the input parameters for creating a short URL.
type ShortenInput struct {
	OriginalURL string // The original long URL to be shortened
}

const (
	// MaxSlugLength defines the maximum length of the generated short code slug.
	MaxSlugLength = 7
)

var generateShortCodeFunc = generateShortCode

// Shorten creates a short URL from the provided original URL.
// If the original URL already exists in the system, it returns the existing short URL.
// Otherwise, it generates a new 7-character short code and creates a new record.
// The operation is idempotent - the same original URL will always return the same short code.
func (i impl) Shorten(ctx context.Context, inp ShortenInput) (model.ShortUrl, error) {
	var err error
	ctx, span := monitoring.Start(ctx, "Controller.Shorten")
	defer monitoring.End(span, &err)

	l := monitoring.Log(ctx)

	// Check if the original URL already has a short code (idempotency check)
	shortUrl, err := i.repo.ShortUrl().GetByOriginalURL(ctx, inp.OriginalURL)
	if err != nil {
		if errors.Is(err, shorturl.ErrNotFound) {
			// URL doesn't exist, generate a new short code and create record
			l.Warn().Msg("[Shorten] shorten URL not found, starting to create")

			return i.createShortURL(ctx, inp)
		}

		l.Error().Stack().Err(err).Msg("[Shorten] shortUrlRepo.GetByOriginalURL err")

		return model.ShortUrl{}, err
	}

	return shortUrl, nil
}

func (i impl) createShortURL(ctx context.Context, inp ShortenInput) (model.ShortUrl, error) {
	var m model.ShortUrl
	var err error
	spanCtx, span := monitoring.Start(ctx, "Repository.DoInTx")
	defer monitoring.End(span, &err)

	if err := i.repo.DoInTx(spanCtx, nil, func(newCtx context.Context, repo repository.Registry) error {
		l := monitoring.Log(newCtx)
		m, err = repo.ShortUrl().Insert(newCtx, model.ShortUrl{
			OriginalURL: inp.OriginalURL,
			Status:      model.ShortUrlStatusActive,
			ShortCode:   generateShortCodeFunc(MaxSlugLength), // Generate random 7-character code
		})

		if err != nil {
			l.Error().Err(err).Msg("[Shorten] ShortUrlRepo.Insert err")
			return err
		}

		_, err = repo.KafkaOutboxEvent().Insert(newCtx, model.KafkaOutboxEvent{
			ID:        id.New(),
			EventType: "urlshortener.metadata.requested.v1",
			Status:    model.KafkaOutboxEventStatusPending,
			Payload: model.Payload{
				EventID:    id.New(),
				OccurredAt: time.Now().UTC(),
				Data: map[string]any{
					"short_code":   m.ShortCode,
					"original_url": m.OriginalURL,
				},
			},
		})
		if err != nil {
			l.Error().Err(err).Msg("[Shorten] KafkaOutboxEventRepo.Insert err")
			return err
		}

		l.Info().Msg("[Shorten] shorten URL created")

		return nil
	}); err != nil {
		return model.ShortUrl{}, pkgerrors.WithStack(err)
	}

	return m, nil
}

// generateShortCode generates a random alphanumeric short code of the specified length.
func generateShortCode(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}

	return string(b)
}
