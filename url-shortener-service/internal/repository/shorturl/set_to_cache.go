package shorturl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	pkgerrors "github.com/pkg/errors"
)

// SetToCache set ShortURL data to cache
func (i impl) SetToCache(ctx context.Context, m model.ShortUrl) error {
	val, err := json.Marshal(m)
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	return pkgerrors.WithStack(i.redisClient.Set(ctx, fmt.Sprintf("%s:%s", cacheKeyShortURL, m.ShortCode), val, cacheShortURLTTL).Err())
}
