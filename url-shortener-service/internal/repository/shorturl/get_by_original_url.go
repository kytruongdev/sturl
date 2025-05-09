package shorturl

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// GetByOriginalURL find short_url record by original_url
func (i impl) GetByOriginalURL(ctx context.Context, originalURL string) (model.ShortUrl, error) {
	o, err := orm.ShortUrls(orm.ShortURLWhere.OriginalURL.EQ(originalURL)).One(ctx, i.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ShortUrl{}, ErrNotFound
		}

		return model.ShortUrl{}, pkgerrors.WithStack(err)
	}

	return toShortUrlModel(*o), nil
}
