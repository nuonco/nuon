package stacks

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/nuonco/nuon/pkg/publicnet"
)

const (
	templateFetchTimeout = 30 * time.Second
	templateDialTimeout  = 10 * time.Second
	maxTemplateRedirects = 3

	// MaxTemplateFetchBytes caps the body of a fetched template or UI definition.
	MaxTemplateFetchBytes = 10 << 20
)

// NewTemplateFetchClient returns the client used to fetch vendor-supplied
// template URLs: nested stack templates and createUiDefinition files. Those
// URLs come straight from app config, so the control plane must not let them
// reach anything but the public internet. The client requires https, dials only
// public addresses (pinning the checked address so a DNS rebind cannot swap it),
// follows a bounded number of https-only redirects, and fails reads past
// MaxTemplateFetchBytes.
//
// Every environment, including local dev, fetches templates from remote
// sources; only tests need NewLocalTemplateFetchClient.
func NewTemplateFetchClient() *http.Client {
	return newTemplateFetchClient(false)
}

// NewLocalTemplateFetchClient is for tests that serve templates from an
// httptest server: it drops the https and public-network checks. The body cap
// still applies. It must not be used outside tests.
func NewLocalTemplateFetchClient() *http.Client {
	return newTemplateFetchClient(true)
}

func newTemplateFetchClient(allowInsecure bool) *http.Client {
	dialer := &net.Dialer{Timeout: templateDialTimeout, KeepAlive: 30 * time.Second}
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = nil
	base.DialContext = publicnet.DialContext(dialer)
	if allowInsecure {
		base.DialContext = dialer.DialContext
	}

	return &http.Client{
		Timeout:   templateFetchTimeout,
		Transport: templateTransport{base: base, allowInsecure: allowInsecure},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > maxTemplateRedirects {
				return errors.New("template fetch: too many redirects")
			}
			req.Header.Del("Referer")
			return validateTemplateURL(req.URL, allowInsecure)
		},
	}
}

type templateTransport struct {
	base          http.RoundTripper
	allowInsecure bool
}

func (t templateTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := validateTemplateURL(req.URL, t.allowInsecure); err != nil {
		return nil, err
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.ContentLength > MaxTemplateFetchBytes {
		resp.Body.Close()
		return nil, fmt.Errorf("template fetch: response exceeds %d bytes", MaxTemplateFetchBytes)
	}
	resp.Body = &cappedBody{ReadCloser: resp.Body, remaining: MaxTemplateFetchBytes}
	return resp, nil
}

func validateTemplateURL(u *url.URL, allowInsecure bool) error {
	if u.Opaque != "" || u.User != nil || u.Hostname() == "" {
		return errors.New("template fetch: URL must be an absolute URL without credentials")
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if allowInsecure {
			return nil
		}
		return errors.New("template fetch: URL must use https")
	default:
		return fmt.Errorf("template fetch: unsupported URL scheme %q", u.Scheme)
	}
}

// cappedBody fails the read once more than the cap has been consumed, unlike
// io.LimitReader which silently truncates and leaves the caller parsing half a
// template.
type cappedBody struct {
	io.ReadCloser
	remaining int64
}

func (b *cappedBody) Read(p []byte) (int, error) {
	if b.remaining <= 0 {
		return 0, fmt.Errorf("template fetch: response exceeds %d bytes", MaxTemplateFetchBytes)
	}
	if int64(len(p)) > b.remaining+1 {
		p = p[:b.remaining+1]
	}
	n, err := b.ReadCloser.Read(p)
	b.remaining -= int64(n)
	if b.remaining < 0 {
		return n, fmt.Errorf("template fetch: response exceeds %d bytes", MaxTemplateFetchBytes)
	}
	return n, err
}
