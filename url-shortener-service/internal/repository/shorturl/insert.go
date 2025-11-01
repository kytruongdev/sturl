package shorturl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"go.opentelemetry.io/otel"
)

// Insert saves data to short_url table
func (i impl) Insert(ctx context.Context, m model.ShortUrl) (model.ShortUrl, error) {
	tracer := otel.Tracer("url-shortener.repository")
	ctx, span := tracer.Start(ctx, "Repository.Insert")
	defer span.End()

	l := logging.FromContext(ctx)
	defer logging.TimeTrack(l, time.Now(), "repository.Insert")

	o := orm.ShortURL{
		ShortCode:   m.ShortCode,
		OriginalURL: m.OriginalURL,
		Status:      m.Status.String(),
	}

	if err := o.Insert(ctx, i.db, boil.Infer()); err != nil {
		span.RecordError(err)
		return model.ShortUrl{}, pkgerrors.WithStack(err)
	}

	m.CreatedAt = o.CreatedAt
	m.UpdatedAt = o.UpdatedAt

	cacheKey := fmt.Sprintf("%s%s", cacheKeyShortURL, m.ShortCode)
	b, err := json.Marshal(m)

	if err != nil {
		span.RecordError(err)
		l.Error().Err(err).Msg("Insert] json.Marshal err")
	}

	rs := i.redisClient.Set(ctx, cacheKey, b, cacheShortURLTTL)
	if rs != nil && rs.Err() != nil {
		span.RecordError(err)
		l.Error().Err(rs.Err()).Msg("[Insert] i.redisClient.Set err")
	}

	l.Info().Str("key", cacheKey).Str("value", string(b)).Msg("[Insert] i.redisClient.Set success")

	return m, nil
}
