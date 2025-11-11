package shorturl

import "errors"

var (
	// ErrInactiveURL means URL is inactive
	ErrInactiveURL = errors.New("URL is inactive")
	// ErrURLNotfound means URL not found
	ErrURLNotfound = errors.New("URL not found")
)
