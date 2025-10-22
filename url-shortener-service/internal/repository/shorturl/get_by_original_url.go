package shorturl

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// GetByOriginalURL find short_url record by original_url
func (i impl) GetByOriginalURL(ctx context.Context, originalURL string) (model.ShortUrl, error) {
	cacheKey := fmt.Sprintf("%s%s", cacheKeyOriginalURL, originalURL)
	// step 1: prioritize fetching data from cache
	val, err := i.redisClient.GetBytes(ctx, cacheKey)
	if err == nil {
		log.Printf("[GetByOriginalURL] i.redisClient.GetBytes by key: %v - result: %+v\n", cacheKey, string(val))
		if val != nil {
			cacheResult := model.ShortUrl{}
			if err = json.Unmarshal(val, &cacheResult); err != nil {
				return cacheResult, pkgerrors.WithStack(err)
			}

			return cacheResult, nil
		}

	}

	// step 2: if data has not stored in cache, get it from database
	log.Println("[GetByOriginalURL] i.redisClient.GetBytes not found, starting get in database")
	o, err := orm.ShortUrls(orm.ShortURLWhere.OriginalURL.EQ(originalURL)).One(ctx, i.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ShortUrl{}, pkgerrors.WithStack(ErrNotFound)
		}

		return model.ShortUrl{}, pkgerrors.WithStack(err)
	}

	// step 3: override data to cache with specific cache key
	m := toShortUrlModel(*o)
	b, err := json.Marshal(m)
	if err != nil {
		log.Printf("[GetByOriginalURL] json marshal error: %+v\n", err)
	}

	rs := i.redisClient.Set(ctx, cacheKey, b, cacheShortURLTTL)
	if rs != nil && rs.Err() != nil {
		log.Printf("[GetByOriginalURL] redis set cache error: %+v\n", err)
	}

	log.Printf("[GetByOriginalURL] i.redisClient.Set success to override new data, key: %+v - value: %+v\n", cacheKey, string(b))

	return m, nil
}
