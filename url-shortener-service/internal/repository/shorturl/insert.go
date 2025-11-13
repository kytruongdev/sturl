package shorturl

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// Insert creates a new short URL record in the database.
func (i impl) Insert(ctx context.Context, m model.ShortUrl) (model.ShortUrl, error) {
	var err error
	ctx, span := monitoring.Start(ctx, "ShortURLRepository.Insert")
	defer monitoring.End(span, &err)

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

	return m, nil
}
