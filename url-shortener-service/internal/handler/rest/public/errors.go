package public

import (
	"net/http"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
)

var (
	// WebErrEmptyOriginalURL means the original_url is empty
	WebErrEmptyOriginalURL = &httpserver.Error{Status: http.StatusBadRequest, Code: "empty_original_url", Desc: "Empty original_url"}
	// WebErrInvalidOriginalURL means the original_url is invalid
	WebErrInvalidOriginalURL = &httpserver.Error{Status: http.StatusBadRequest, Code: "invalid_original_url", Desc: "Invalid original_url"}
)

func convertControllerError(err error) error {
	// TODO: implement detail here
	switch err {
	default:
		return err
	}
}
