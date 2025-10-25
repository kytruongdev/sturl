package validator

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	pkgerrors "github.com/pkg/errors"
)

// ValidateURL validate url is valid or not
func ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("unsupported scheme")
	}

	if _, err = net.LookupHost(u.Hostname()); err != nil {
		return pkgerrors.WithStack(err)
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Prevent following redirects
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Head(rawURL)
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	resp.Body.Close()

	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		return errors.New("invalid path")
	}

	return nil
}
