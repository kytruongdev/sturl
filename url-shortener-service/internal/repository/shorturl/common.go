package shorturl

import (
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
)

const (
	// cacheKeyShortURL is the Redis key prefix for caching short URLs by short code.
	cacheKeyShortURL = "short_url:"
	// cacheKeyOriginalURL is the Redis key prefix for caching short URLs by original URL.
	cacheKeyOriginalURL = "original_url:"
	// cacheShortURLTTL is the time-to-live duration for cached short URL entries (24 hours).
	cacheShortURLTTL = 24 * time.Hour
)

func toShortUrlModel(o orm.ShortURL) model.ShortUrl {
	return model.ShortUrl{
		ShortCode:   o.ShortCode,
		OriginalURL: o.OriginalURL,
		Status:      model.ShortUrlStatus(o.Status),
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}
