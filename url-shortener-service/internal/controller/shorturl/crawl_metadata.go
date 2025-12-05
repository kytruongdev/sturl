package shorturl

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// CrawlURLMetadata crawls metadata from the original URL, save to database and emits an outgoing event.
func (i impl) CrawlURLMetadata(ctx context.Context, shortCode string) (model.UrlMetadata, error) {
	var err error
	ctx, span := monitoring.Start(ctx, "ShortURLController.CrawlURLMetadata")
	defer monitoring.End(span, &err)

	log := monitoring.Log(ctx).Field("short_code", shortCode)
	log.Info().Msg("[CrawlURLMetadata] starting crawling url...")

	su, err := i.repo.ShortUrl().GetByShortCode(ctx, shortCode)
	if err != nil {
		log.Error().Err(err).Msg("[CrawlMetadata] shortUrlRepo.GetByShortCode err")
		if errors.Is(err, sql.ErrNoRows) {
			return model.UrlMetadata{}, ErrURLNotfound
		}
		return model.UrlMetadata{}, err
	}

	crawledMetadata, err := newURLMetadataCrawler().crawl(ctx, su.OriginalURL)
	if err != nil {
		log.Error().Err(err).Msg("[CrawlMetadata] i.crawl err")
		return model.UrlMetadata{}, err
	}

	log.
		Field("title", crawledMetadata.Title).
		Field("description", crawledMetadata.Description).
		Field("image", crawledMetadata.Image).
		Field("favicon", crawledMetadata.Favicon).
		Info().Msg("[CrawlURLMetadata] crawler completed")

	if err = i.repo.DoInTx(ctx, nil, func(txCtx context.Context, txRepo repository.Registry) error {
		txLog := monitoring.Log(txCtx)
		txLog.Info().Msg("[CrawlURLMetadata] starting DoInTx")

		if err = i.updateMetadata(txCtx, txRepo, crawledMetadata, su.ShortCode); err != nil {
			txLog.Error().Err(err).Msg("[DoInTx] updateMetadata err")
			return err
		}

		meta := monitoring.SpanMetadataFromContext(txCtx)
		if err = i.insertOutgoingEvent(txCtx, txRepo, model.OutgoingEvent{
			ID:            newIDFunc(),
			Topic:         model.TopicMetadataCrawledV1,
			Status:        model.OutgoingEventStatusPending,
			CorrelationID: meta.CorrelationID,
			TraceID:       meta.TraceID,
			SpanID:        meta.SpanID,
			Payload: model.Payload{
				EventID:    newIDFunc(),
				OccurredAt: time.Now().UTC(),
				Data: map[string]string{
					"short_code":   su.ShortCode,
					"original_url": su.OriginalURL,
				},
			},
		}); err != nil {
			txLog.Error().Err(err).Msg("[DoInTx] insertOutgoingEvent err")
			return err
		}

		txLog.Info().Msg("[CrawlURLMetadata] DoInTx successfully completed")
		return nil
	}); err != nil {
		return model.UrlMetadata{}, err
	}

	return crawledMetadata, nil
}

func (i impl) updateMetadata(ctx context.Context, txRepo repository.Registry, metadata model.UrlMetadata, shortCode string) error {
	return txRepo.ShortUrl().Update(ctx, model.ShortUrl{
		Metadata: metadata,
	}, shortCode)
}

func (i impl) insertOutgoingEvent(ctx context.Context, txRepo repository.Registry, event model.OutgoingEvent) error {
	_, err := txRepo.OutgoingEvent().Insert(ctx, event)
	return err
}

type urlMetadataCrawler struct {
	client http.Client
}

func newURLMetadataCrawler() urlMetadataCrawler {
	return urlMetadataCrawler{
		client: http.Client{
			Timeout: 5 * time.Second, // Prevent long-hanging crawls
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Reject excessive redirects to avoid loops
				if len(via) >= 5 {
					return errors.New("too many redirects")
				}
				return nil
			},
		},
	}
}

// crawl fetches HTML head → parse metadata → build result
func (i urlMetadataCrawler) crawl(ctx context.Context, rawURL string) (model.UrlMetadata, error) {
	rawURL = upgradeToHTTPS(rawURL) // Auto-upgrade http→https for reliability

	body, baseURL, err := i.fetchHeadHTML(ctx, rawURL)
	if err != nil {
		return model.UrlMetadata{}, err
	}

	// parse <head> to extract meta tags, title, og:*, favicon
	head := parseHeadMetadata(body)

	// build final strongly-typed metadata struct
	return buildMetadata(head, baseURL), nil
}

// fetchHeadHTML fetches only the HEAD portion of the HTML (limited bytes) for faster crawling.
// Many sites place metadata within the first ~100KB of HTML.
func (i urlMetadataCrawler) fetchHeadHTML(ctx context.Context, rawURL string) ([]byte, *url.URL, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, nil, err
	}

	// Pretend to be a normal desktop Chrome browser
	const defaultUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/129.0 Safari/537.36"

	req.Header.Set("User-Agent", defaultUA)

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// Reject non-success HTTP codes
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, rawURL)
	}

	// Read-only the first N bytes, usually enough to include <head>
	const headReadLimit = 256 * 1024
	limited := io.LimitReader(resp.Body, headReadLimit)

	// Attempt charset decoding when possible (UTF-8, windows-1258, etc.)
	contentType := resp.Header.Get("Content-Type")
	reader, err := charset.NewReader(limited, contentType)
	if err != nil {
		// Fallback: read raw bytes if charset conversion fails
		limited = io.LimitReader(resp.Body, headReadLimit)
		reader = limited
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, err
	}

	baseURL, _ := url.Parse(resp.Request.URL.String())
	return body, baseURL, nil
}

// headMeta is internal model storing extracted <head> metadata
type headMeta struct {
	Title       string
	Description string
	OgTitle     string
	OgDesc      string
	OgImage     string
	Favicon     string
}

func parseHeadMetadata(body []byte) headMeta {
	z := html.NewTokenizer(bytes.NewReader(body))
	h := headMeta{}
	var inTitle bool

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		tok := z.Token()
		switch tt {

		case html.StartTagToken, html.SelfClosingTagToken:
			if tok.Data == "title" {
				inTitle = true
			}

			if tok.Data == "meta" {
				parseMetaTag(tok, &h)
			}

			if tok.Data == "link" {
				parseFavicon(tok, &h)
			}

		case html.TextToken:
			// Capture <title>text</title>
			if inTitle && h.Title == "" {
				h.Title = strings.TrimSpace(tok.Data)
			}

		case html.EndTagToken:
			if tok.Data == "title" {
				inTitle = false
			}
			// Stop parsing once </head> is reached; body is irrelevant
			if tok.Data == "head" {
				return h
			}
		}
	}

	return h
}

func buildMetadata(h headMeta, base *url.URL) model.UrlMetadata {
	return model.UrlMetadata{
		FinalURL:    base.String(),
		Title:       firstNonEmpty(h.OgTitle, h.Title),
		Description: firstNonEmpty(h.OgDesc, h.Description),
		Image:       resolveURL(h.OgImage, base),
		Favicon:     resolveURL(h.Favicon, base),
	}
}

// Extract <meta> tags such as:
//
//	<meta property="og:title" content="...">
//	<meta name="description" content="...">
func parseMetaTag(tok html.Token, h *headMeta) {
	var name, prop, content string
	for _, a := range tok.Attr {
		switch strings.ToLower(a.Key) {
		case "name":
			name = strings.ToLower(a.Val)
		case "property":
			prop = strings.ToLower(a.Val)
		case "content":
			content = a.Val
		}
	}

	if prop == "og:title" {
		h.OgTitle = content
	}
	if prop == "og:description" {
		h.OgDesc = content
	}
	if prop == "og:image" {
		h.OgImage = content
	}
	if name == "description" {
		h.Description = content
	}
}

// Extract favicon from:
//
//	<link rel="icon" href="...">
//	<link rel="shortcut icon" href="...">
func parseFavicon(tok html.Token, h *headMeta) {
	var rel, href string
	for _, a := range tok.Attr {
		switch strings.ToLower(a.Key) {
		case "rel":
			rel = a.Val
		case "href":
			href = a.Val
		}
	}

	if rel == "icon" || rel == "shortcut icon" {
		h.Favicon = href
	}
}

// Upgrade http:// links to https:// for better security and reliability
func upgradeToHTTPS(url string) string {
	if strings.HasPrefix(url, "http://") {
		return "https://" + strings.TrimPrefix(url, "http://")
	}
	return url
}

// Utility selecting the first non-empty string
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// Resolve relative URLs against base:
//
//	"/image.jpg" → "https://example.com/image.jpg"
func resolveURL(resource string, base *url.URL) string {
	if resource == "" {
		return ""
	}

	u, err := url.Parse(resource)
	if err != nil {
		return resource
	}

	if u.IsAbs() {
		return u.String()
	}

	return base.ResolveReference(u).String()
}
