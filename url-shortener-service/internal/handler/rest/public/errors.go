package public

import (
	"net/http"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
)

var (
	// WebErrEmptyShortCode means the short code is empty
	WebErrEmptyShortCode = &httpserver.Error{Status: http.StatusBadRequest, Code: "empty_short_code", Desc: "Empty short_code"}
	// WebErrEmptyOriginalURL means the original_url is empty
	WebErrEmptyOriginalURL = &httpserver.Error{Status: http.StatusBadRequest, Code: "empty_original_url", Desc: "URL is empty"}
	// WebErrInvalidOriginalURL means the original_url is invalid
	WebErrInvalidOriginalURL = &httpserver.Error{Status: http.StatusBadRequest, Code: "invalid_original_url", Desc: "URL is invalid"}
	// WebErrInactiveOriginalURL means the original_url is invalid
	WebErrInactiveOriginalURL = &httpserver.Error{Status: http.StatusBadRequest, Code: "inactive_url", Desc: "URL is inactive"}
)

func convertControllerError(err error) error {
	switch err {
	case shorturl.ErrInactiveURL:
		return WebErrInactiveOriginalURL
	default:
		return err
	}
}
