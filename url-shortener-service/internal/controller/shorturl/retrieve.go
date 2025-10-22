package shorturl

import (
	"context"
	"log"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
)

// Retrieve retrieves short url by short code
func (i impl) Retrieve(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	m, err := i.shortUrlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		log.Printf("[Retrieve] shortUrlRepo.GetByShortCode err: %+v\n", err)
		return model.ShortUrl{}, err
	}

	if m.Status != model.ShortUrlStatusActive {
		return model.ShortUrl{}, ErrInactiveURL
	}

	return m, err
}
