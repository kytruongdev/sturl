package shorturl

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"

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
	shortUrl, err := i.shortUrlRepo.GetByOriginalURL(ctx, inp.OriginalURL)
	if err != nil {
		log.Printf("[Shorten] shortUrlRepo.GetByOriginalURL err: %+v\n", err)

		if errors.Is(err, shorturl.ErrNotFound) {
			log.Println("[Shorten] shorten URL not found, starting to create")
			m, err := i.shortUrlRepo.Insert(ctx, model.ShortUrl{
				OriginalURL: inp.OriginalURL,
				Status:      model.ShortUrlStatusActive,
				ShortCode:   generateShortCodeFunc(MaxSlugLength),
			})

			if err != nil {
				log.Printf("[Shorten] shortUrlRepo.Insert err: %+v\n", err)
				return model.ShortUrl{}, err
			}

			log.Println("[Shorten] shorten URL created")

			return m, nil
		}

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
