package shorturl

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/id"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/outgoingevent"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCrawlURLMetadata(t *testing.T) {
	tcs := map[string]struct {
		shortCode                   string
		mockGetByShortCodeWant      model.ShortUrl
		mockGetByShortCodeErr       error
		mockUpdateErr               error
		mockInsertOutgoingEventWant model.OutgoingEvent
		mockInsertOutgoingEventErr  error
		wantErr                     error
	}{
		"success - crawl and save metadata": {
			shortCode: "abc123",
			mockGetByShortCodeWant: model.ShortUrl{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				Status:      model.ShortUrlStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockGetByShortCodeErr: nil,
			mockUpdateErr:         nil,
			mockInsertOutgoingEventWant: model.OutgoingEvent{
				ID:     123,
				Topic:  model.TopicMetadataCrawledV1,
				Status: model.OutgoingEventStatusPending,
			},
			mockInsertOutgoingEventErr: nil,
			wantErr:                    nil,
		},

		"fail - short code not found": {
			shortCode:             "notfound",
			mockGetByShortCodeErr: sql.ErrNoRows,
			wantErr:               ErrURLNotfound,
		},

		"fail - GetByShortCode returns error": {
			shortCode:             "abc123",
			mockGetByShortCodeErr: errors.New("database error"),
			wantErr:               errors.New("database error"),
		},

		"fail - update metadata fails": {
			shortCode: "abc123",
			mockGetByShortCodeWant: model.ShortUrl{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				Status:      model.ShortUrlStatusActive,
			},
			mockGetByShortCodeErr: nil,
			mockUpdateErr:         errors.New("update failed"),
			wantErr:               errors.New("update failed"),
		},

		"fail - insert outgoing event fails": {
			shortCode: "abc123",
			mockGetByShortCodeWant: model.ShortUrl{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				Status:      model.ShortUrlStatusActive,
			},
			mockGetByShortCodeErr:      nil,
			mockUpdateErr:              nil,
			mockInsertOutgoingEventErr: errors.New("outbox insert failed"),
			wantErr:                    errors.New("outbox insert failed"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			// Mock ID generator
			newIDFunc = func() int64 {
				return 123
			}
			defer func() { newIDFunc = id.New }()

			// Mock ShortURL repo
			mockShort := new(shorturl.MockRepository)
			mockShort.On("GetByShortCode", mock.Anything, tc.shortCode).
				Return(tc.mockGetByShortCodeWant, tc.mockGetByShortCodeErr)

			if tc.mockGetByShortCodeErr == nil {
				mockShort.On("Update", mock.Anything, mock.Anything, tc.shortCode).
					Return(tc.mockUpdateErr)
			}

			// Mock Outbox repo
			mockOutbox := new(outgoingevent.MockRepository)
			if tc.mockGetByShortCodeErr == nil && tc.mockUpdateErr == nil {
				mockOutbox.On("Insert", mock.Anything, mock.Anything).
					Return(tc.mockInsertOutgoingEventWant, tc.mockInsertOutgoingEventErr)
			}

			// Mock Registry
			mockReg := new(repository.MockRegistry)
			mockReg.On("ShortUrl").Return(mockShort)
			mockReg.On("OutgoingEvent").Return(mockOutbox)

			// Fake DoInTx: simply run fn
			mockReg.On("DoInTx", mock.Anything, mock.Anything, mock.Anything).
				Return(func(ctx context.Context, _ backoff.BackOff, fn func(context.Context, repository.Registry) error) error {
					return fn(ctx, mockReg)
				})

			i := New(mockReg)

			_, err := i.CrawlURLMetadata(ctx, tc.shortCode)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUpgradeToHTTPS(t *testing.T) {
	tcs := map[string]struct {
		input string
		want  string
	}{
		"http to https": {
			input: "http://example.com",
			want:  "https://example.com",
		},
		"already https": {
			input: "https://example.com",
			want:  "https://example.com",
		},
		"no protocol": {
			input: "example.com",
			want:  "example.com",
		},
		"http with path": {
			input: "http://example.com/path/to/page",
			want:  "https://example.com/path/to/page",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			actual := upgradeToHTTPS(tc.input)
			require.Equal(t, tc.want, actual)
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tcs := map[string]struct {
		values []string
		want   string
	}{
		"first value is non-empty": {
			values: []string{"first", "second", "third"},
			want:   "first",
		},
		"first value is empty, second is not": {
			values: []string{"", "second", "third"},
			want:   "second",
		},
		"all empty": {
			values: []string{"", "", ""},
			want:   "",
		},
		"only whitespace values": {
			values: []string{"  ", "\t", "\n"},
			want:   "",
		},
		"mixed empty and whitespace": {
			values: []string{"", "  ", "valid"},
			want:   "valid",
		},
		"no values": {
			values: []string{},
			want:   "",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			actual := firstNonEmpty(tc.values...)
			require.Equal(t, tc.want, actual)
		})
	}
}

func TestResolveURL(t *testing.T) {
	tcs := map[string]struct {
		resource string
		baseURL  string
		want     string
	}{
		"absolute URL": {
			resource: "https://cdn.example.com/image.jpg",
			baseURL:  "https://example.com",
			want:     "https://cdn.example.com/image.jpg",
		},
		"relative path": {
			resource: "/image.jpg",
			baseURL:  "https://example.com",
			want:     "https://example.com/image.jpg",
		},
		"relative path with subdirectory": {
			resource: "/assets/image.jpg",
			baseURL:  "https://example.com/page",
			want:     "https://example.com/assets/image.jpg",
		},
		"relative without leading slash": {
			resource: "image.jpg",
			baseURL:  "https://example.com/page/",
			want:     "https://example.com/page/image.jpg",
		},
		"empty resource": {
			resource: "",
			baseURL:  "https://example.com",
			want:     "",
		},
		"protocol-relative URL": {
			resource: "//cdn.example.com/image.jpg",
			baseURL:  "https://example.com",
			want:     "https://cdn.example.com/image.jpg",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			// Parse base URL
			base, err := parseURL(tc.baseURL)
			require.NoError(t, err)

			actual := resolveURL(tc.resource, base)
			require.Equal(t, tc.want, actual)
		})
	}
}

// Helper function for tests
func parseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

func TestParseHeadMetadata(t *testing.T) {
	tcs := map[string]struct {
		htmlBody string
		want     headMeta
	}{
		"complete metadata": {
			htmlBody: `
<!DOCTYPE html>
<html>
<head>
	<title>Example Title</title>
	<meta name="description" content="Example description">
	<meta property="og:title" content="OG Title">
	<meta property="og:description" content="OG Description">
	<meta property="og:image" content="https://example.com/image.jpg">
	<link rel="icon" href="/favicon.ico">
</head>
<body>Content</body>
</html>`,
			want: headMeta{
				Title:       "Example Title",
				Description: "Example description",
				OgTitle:     "OG Title",
				OgDesc:      "OG Description",
				OgImage:     "https://example.com/image.jpg",
				Favicon:     "/favicon.ico",
			},
		},
		"minimal metadata": {
			htmlBody: `
<!DOCTYPE html>
<html>
<head>
	<title>Simple Title</title>
</head>
<body>Content</body>
</html>`,
			want: headMeta{
				Title:       "Simple Title",
				Description: "",
				OgTitle:     "",
				OgDesc:      "",
				OgImage:     "",
				Favicon:     "",
			},
		},
		"og tags only": {
			htmlBody: `
<!DOCTYPE html>
<html>
<head>
	<meta property="og:title" content="Only OG">
	<meta property="og:description" content="Only OG Desc">
</head>
<body>Content</body>
</html>`,
			want: headMeta{
				Title:       "",
				Description: "",
				OgTitle:     "Only OG",
				OgDesc:      "Only OG Desc",
				OgImage:     "",
				Favicon:     "",
			},
		},
		"shortcut icon": {
			htmlBody: `
<!DOCTYPE html>
<html>
<head>
	<link rel="shortcut icon" href="/favicon.png">
</head>
<body>Content</body>
</html>`,
			want: headMeta{
				Title:       "",
				Description: "",
				OgTitle:     "",
				OgDesc:      "",
				OgImage:     "",
				Favicon:     "/favicon.png",
			},
		},
		"empty html": {
			htmlBody: ``,
			want: headMeta{
				Title:       "",
				Description: "",
				OgTitle:     "",
				OgDesc:      "",
				OgImage:     "",
				Favicon:     "",
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			actual := parseHeadMetadata([]byte(tc.htmlBody))
			require.True(t,
				cmp.Equal(tc.want, actual),
				"diff: %v",
				cmp.Diff(tc.want, actual),
			)
		})
	}
}

func TestBuildMetadata(t *testing.T) {
	tcs := map[string]struct {
		head    headMeta
		baseURL string
		want    model.UrlMetadata
	}{
		"prefers og tags over regular tags": {
			head: headMeta{
				Title:       "Regular Title",
				Description: "Regular Description",
				OgTitle:     "OG Title",
				OgDesc:      "OG Description",
				OgImage:     "/image.jpg",
				Favicon:     "/favicon.ico",
			},
			baseURL: "https://example.com",
			want: model.UrlMetadata{
				FinalURL:    "https://example.com",
				Title:       "OG Title",
				Description: "OG Description",
				Image:       "https://example.com/image.jpg",
				Favicon:     "https://example.com/favicon.ico",
			},
		},
		"fallback to regular tags when og tags empty": {
			head: headMeta{
				Title:       "Regular Title",
				Description: "Regular Description",
				OgTitle:     "",
				OgDesc:      "",
				OgImage:     "",
				Favicon:     "/favicon.ico",
			},
			baseURL: "https://example.com",
			want: model.UrlMetadata{
				FinalURL:    "https://example.com",
				Title:       "Regular Title",
				Description: "Regular Description",
				Image:       "",
				Favicon:     "https://example.com/favicon.ico",
			},
		},
		"absolute URLs preserved": {
			head: headMeta{
				OgImage: "https://cdn.example.com/image.jpg",
				Favicon: "https://cdn.example.com/favicon.ico",
			},
			baseURL: "https://example.com",
			want: model.UrlMetadata{
				FinalURL:    "https://example.com",
				Title:       "",
				Description: "",
				Image:       "https://cdn.example.com/image.jpg",
				Favicon:     "https://cdn.example.com/favicon.ico",
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			base, err := parseURL(tc.baseURL)
			require.NoError(t, err)

			actual := buildMetadata(tc.head, base)
			require.True(t,
				cmp.Equal(tc.want, actual),
				"diff: %v",
				cmp.Diff(tc.want, actual),
			)
		})
	}
}
