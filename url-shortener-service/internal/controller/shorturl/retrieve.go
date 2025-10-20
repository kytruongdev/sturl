package shorturl

import (
	"context"
	"log"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
)

// Retrieve retrieves short url by short code
func (i impl) Retrieve(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	cacheData, err := i.shortUrlRepo.GetByShortCodeFromCache(ctx, shortCode)
	if err == nil {
		return cacheData, nil
	}

	log.Printf("Failed to get shortcode [%s] from cache, err: %v\n", shortCode, err)

	m, err := i.shortUrlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return model.ShortUrl{}, err
	}

	if m.Status != model.ShortUrlStatusActive {
		return model.ShortUrl{}, ErrInactiveURL
	}

	i.setToCacheSafe(ctx, m)

	log.Printf("Successfully set shortcode [%s] to cache\n", shortCode)

	return m, err
}

func (i impl) setToCacheSafe(ctx context.Context, m model.ShortUrl) {
	go func(ctx context.Context, shortURL model.ShortUrl) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic in setToCacheSafe:", r)
			}
		}()

		const duration = 2 * time.Second
		newCtx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()

		if err := i.shortUrlRepo.SetToCache(newCtx, shortURL); err != nil {
			log.Printf("Failed to set shortcode [%s] to cache, err: %v\n", m.ShortCode, err)
		}
	}(ctx, m)
}
