package shorturl

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// GetByShortCode retrieves a short URL record by its short code using a cache-aside pattern.
// It first checks Redis cache, and if not found, queries the database and updates the cache.
// This optimizes read performance by reducing database load for frequently accessed URLs.
func (i impl) GetByShortCode(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	var err error
	ctx, span := monitoring.Start(ctx, "Repository.GetByShortCode")
	defer monitoring.End(span, &err)

	l := monitoring.Log(ctx)

	cacheKey := fmt.Sprintf("%s%s", cacheKeyShortURL, shortCode)
	
	// Step 1: Try to fetch from Redis cache first (cache-aside pattern)
	val, err := i.redisClient.GetBytes(ctx, cacheKey)
	if err == nil {
		l.Info().Str("cacheKey", cacheKey).Msgf("[GetByShortCode] i.redisClient.GetBytes - result: %+v\n", string(val))

		if val != nil {
			// Cache hit - deserialize and return
			cacheResult := model.ShortUrl{}
			if err = json.Unmarshal(val, &cacheResult); err != nil {
				return cacheResult, pkgerrors.WithStack(err)
			}

			return cacheResult, nil
		}

	}

	// Step 2: Cache miss - fetch from database
	l.Warn().Msg("[GetByShortCode] i.redisClient.GetBytes not found, starting get in database")
	o, err := orm.FindShortURL(ctx, i.db, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ShortUrl{}, pkgerrors.WithStack(ErrNotFound)
		}

		return model.ShortUrl{}, pkgerrors.WithStack(err)
	}

	// Step 3: Update cache with the fetched data for future requests
	m := toShortUrlModel(*o)
	val, err = json.Marshal(m)
	if err != nil {
		l.Error().Err(err).Msg("[GetByShortCode] json.Marshal err")
	}

	// Cache update is best-effort - errors are logged but don't fail the operation
	rs := i.redisClient.Set(ctx, cacheKey, val, cacheShortURLTTL)
	if rs != nil && rs.Err() != nil {
		l.Error().Err(rs.Err()).Msg("[GetByShortCode] i.redisClient.Set err")
	}

	return m, nil
}
