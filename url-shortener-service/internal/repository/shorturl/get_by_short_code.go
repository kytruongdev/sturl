package shorturl

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// GetByShorCode find short_url record by short_code
func (i impl) GetByShorCode(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	o, err := orm.FindShortURL(ctx, i.db, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ShortUrl{}, pkgerrors.WithStack(ErrNotFound)
		}

		return model.ShortUrl{}, pkgerrors.WithStack(err)
	}

	return toShortUrlModel(*o), nil
}
