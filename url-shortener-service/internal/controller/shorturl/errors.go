package shorturl

import "errors"

var (
	// ErrInactiveURL means URL is inactive
	ErrInactiveURL = errors.New("URL is inactive")
)
