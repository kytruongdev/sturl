package shorturl

import (
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
)

func toShortUrlModel(o orm.ShortURL) model.ShortUrl {
	return model.ShortUrl{
		ShortCode:   o.ShortCode,
		OriginalURL: o.OriginalURL,
		Status:      model.ShortUrlStatus(o.Status),
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}
