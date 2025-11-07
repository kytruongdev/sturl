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

// Insert saves data to short_url table
func (i impl) Insert(ctx context.Context, m model.ShortUrl) (model.ShortUrl, error) {
	var err error
	monitor := monitoring.FromContext(ctx)
	ctx, span, l := monitor.StartSpanWithLog(ctx, "Repository.Insert")
	defer monitor.EndSpan(&span, &err)

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

	cacheKey := fmt.Sprintf("%s%s", cacheKeyShortURL, m.ShortCode)
	b, err := json.Marshal(m)

	if err != nil {
		l.Error().Err(err).Msg("Insert] json.Marshal err")
	}

	rs := i.redisClient.Set(ctx, cacheKey, b, cacheShortURLTTL)
	if rs != nil && rs.Err() != nil {
		l.Error().Err(rs.Err()).Msg("[Insert] i.redisClient.Set err")
	}

	l.Info().Str("key", cacheKey).Str("value", string(b)).Msg("[Insert] i.redisClient.Set success")

	return m, nil
}
