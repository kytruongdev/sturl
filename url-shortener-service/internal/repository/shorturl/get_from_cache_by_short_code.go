package shorturl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	pkgerrors "github.com/pkg/errors"
)

// GetByShortCodeFromCache get ShortURL data model from cache by shortCode
func (i impl) GetByShortCodeFromCache(ctx context.Context, shortCode string) (model.ShortUrl, error) {
	val, err := i.redisClient.Get(ctx, fmt.Sprintf("%s:%s", cacheKeyShortURL, shortCode)).Result()
	if err != nil {
		return model.ShortUrl{}, pkgerrors.WithStack(err)
	}

	var cacheModel model.ShortUrl
	if err = json.Unmarshal([]byte(val), &cacheModel); err != nil {
		return model.ShortUrl{}, pkgerrors.WithStack(err)
	}

	return cacheModel, nil
}
