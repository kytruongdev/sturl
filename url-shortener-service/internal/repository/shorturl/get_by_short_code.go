package shorturl

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// GetByShortCode find short_url record by short_code
func (i impl) GetByShortCode(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	l := logging.FromContext(ctx)
	defer logging.TimeTrack(l, time.Now(), "repository.GetByShortCode")

	cacheKey := fmt.Sprintf("%s%s", cacheKeyShortURL, shortCode)
	// step 1: prioritize fetching data from cache
	val, err := i.redisClient.GetBytes(ctx, cacheKey)
	if err == nil {
		l.Info().Str("cacheKey", cacheKey).Msgf("[GetByShortCode] i.redisClient.GetBytes - result: %+v\n", string(val))
		if val != nil {
			cacheResult := model.ShortUrl{}
			err = json.Unmarshal(val, &cacheResult)
			return cacheResult, pkgerrors.WithStack(err)
		}

	}

	// step 2: if data has not stored in cache, get it from database
	l.Warn().Msg("[GetByShortCode] i.redisClient.GetBytes not found, starting get in database")
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
		l.Error().Err(err).Msg("[GetByShortCode] json.Marshal err")
	}

	rs := i.redisClient.Set(ctx, cacheKey, val, cacheShortURLTTL)
	if rs != nil && rs.Err() != nil {
		l.Error().Err(rs.Err()).Msg("[GetByShortCode] i.redisClient.Set err")
	}

	return m, nil
}
