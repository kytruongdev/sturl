package shorturl

import (
	"context"
	"encoding/json"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

func (i impl) Update(ctx context.Context, m model.ShortUrl, shortCode string) error {
	var err error
	ctx, span := monitoring.Start(ctx, "ShortUrlRepository.Update")
	defer monitoring.End(span, &err)

	current, err := orm.FindShortURL(ctx, i.db, shortCode)
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	whitelist := []string{
		orm.ShortURLColumns.OriginalURL,
	}

	if m.Status != "" {
		current.Status = m.Status.String()
		whitelist = append(whitelist, orm.ShortURLColumns.Status)
	}

	if m.Metadata.IsNotEmpty() {
		b, err := json.Marshal(m.Metadata)
		if err != nil {
			return pkgerrors.WithStack(err)
		}
		current.Metadata = null.JSONFrom(b)
		whitelist = append(whitelist, orm.ShortURLColumns.Metadata)
	}

	_, err = current.Update(ctx, i.db, boil.Whitelist(whitelist...))
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	return nil
}
