package model

import "time"

// ShortUrlStatus represents status of `short_url`
type ShortUrlStatus string

const (
	// ShortUrlStatusActive means `short_url` is active
	ShortUrlStatusActive ShortUrlStatus = "ACTIVE"
	// ShortUrlStatusInactive means `short_url` is inactive
	ShortUrlStatusInactive ShortUrlStatus = "INACTIVE"
	// ShortUrlStatusDeleted means `short_url` is deleted
	ShortUrlStatusDeleted ShortUrlStatus = "DELETED"
)

// String converts to string value
func (stt ShortUrlStatus) String() string {
	return string(stt)
}

// IsValid checks if short_url status is valid
func (stt ShortUrlStatus) IsValid() bool {
	return stt == ShortUrlStatusActive || stt == ShortUrlStatusInactive || stt == ShortUrlStatusDeleted
}

// ShortUrl represents business model of `short_url`
type ShortUrl struct {
	ShortCode   string
	OriginalURL string
	Status      ShortUrlStatus
	Metadata    UrlMetadata
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
