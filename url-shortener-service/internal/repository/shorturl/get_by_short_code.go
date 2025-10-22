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

// GetByShortCode find short_url record by short_code
func (i impl) GetByShortCode(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	cacheKey := fmt.Sprintf("%s%s", cacheKeyShortURL, shortCode)
	// step 1: prioritize fetching data from cache
	val, err := i.redisClient.GetBytes(ctx, cacheKey)
	if err == nil {
		log.Printf("[GetByShortCode] i.redisClient.GetBytes by key: %v - result: %+v\n", cacheKey, string(val))
		if val != nil {
			cacheResult := model.ShortUrl{}
			err = json.Unmarshal(val, &cacheResult)
			return cacheResult, pkgerrors.WithStack(err)
		}

	}

	// step 2: if data has not stored in cache, get it from database
	log.Println("[GetByShortCode] i.redisClient.GetBytes not found, starting get in database")
	o, err := orm.FindShortURL(ctx, i.db, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ShortUrl{}, pkgerrors.WithStack(ErrNotFound)
		}

		return model.ShortUrl{}, pkgerrors.WithStack(err)
	}

	// step 3: override data to cache with specific cache key
	m := toShortUrlModel(*o)
	val, err = json.Marshal(m)
	if err != nil {
		log.Printf("[GetByShortCode] json marshal error: %v\n", err)
	}

	rs := i.redisClient.Set(ctx, cacheKey, val, cacheShortURLTTL)
	if rs != nil && rs.Err() != nil {
		log.Printf("[GetByShortCode] redis set error: %v\n", err)
	}

	log.Printf("[GetByShortCode] i.redisClient.Set success to override new data, key: %+v - value: %+v\n", cacheKey, string(val))

	return m, nil
}
