package shorturl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
)

// Insert creates a new short URL record in the database and updates the cache.
// It saves the short URL to PostgreSQL and then caches it in Redis for fast retrieval.
// Note: Cache update failures are logged but don't cause the operation to fail.
func (i impl) Insert(ctx context.Context, m model.ShortUrl) (model.ShortUrl, error) {
	var err error
	ctx, span := monitoring.Start(ctx, "Repository.GetByOriginalURL")
	defer monitoring.End(span, &err)

	l := monitoring.Log(ctx)

	o := orm.ShortURL{
		ShortCode:   m.ShortCode,
		OriginalURL: m.OriginalURL,
		Status:      m.Status.String(),
	}

	if err := o.Insert(ctx, i.db, boil.Infer()); err != nil {
		return model.ShortUrl{}, pkgerrors.WithStack(err)
	}

	m.CreatedAt = o.CreatedAt
	m.UpdatedAt = o.UpdatedAt

	// Update cache after successful database insert (write-through pattern)
	cacheKey := fmt.Sprintf("%s%s", cacheKeyShortURL, m.ShortCode)
	b, err := json.Marshal(m)

	if err != nil {
		l.Error().Err(err).Msg("Insert] json.Marshal err")
	}

	// Cache update is best-effort - errors are logged but don't fail the operation
	rs := i.redisClient.Set(ctx, cacheKey, b, cacheShortURLTTL)
	if rs != nil && rs.Err() != nil {
		l.Error().Err(rs.Err()).Msg("[Insert] i.redisClient.Set err")
	}

	l.Info().Str("key", cacheKey).Str("value", string(b)).Msg("[Insert] i.redisClient.Set success")

	return m, nil
}
