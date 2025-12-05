package shorturl

import (
	"encoding/json"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

const (
	// cacheKeyShortURL is the Redis key prefix for caching short URLs by short code.
	cacheKeyShortURL = "short_url:"
	// cacheKeyOriginalURL is the Redis key prefix for caching short URLs by original URL.
	cacheKeyOriginalURL = "original_url:"
	// cacheShortURLTTL is the time-to-live duration for cached short URL entries (24 hours).
	cacheShortURLTTL = 24 * time.Hour
)

func toShortUrlModel(o orm.ShortURL) (model.ShortUrl, error) {
	var metadata model.UrlMetadata
	if o.Metadata.Valid {
		if err := json.Unmarshal(o.Metadata.JSON, &metadata); err != nil {
			return model.ShortUrl{}, pkgerrors.WithStack(err)
		}
	}

	return model.ShortUrl{
		ShortCode:   o.ShortCode,
		OriginalURL: o.OriginalURL,
		Status:      model.ShortUrlStatus(o.Status),
		Metadata:    metadata,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}, nil
}
