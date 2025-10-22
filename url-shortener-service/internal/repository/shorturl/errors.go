package shorturl

import "errors"

var (
	// ErrNotFound means no short_url record found
	ErrNotFound = errors.New("short_url record not found")
)
