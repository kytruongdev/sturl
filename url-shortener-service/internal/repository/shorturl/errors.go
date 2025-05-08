package shorturl

import "errors"

var (
	// ErrNotFound means no short_url record found
	ErrNotFound = errors.New("short validator not found")
)
