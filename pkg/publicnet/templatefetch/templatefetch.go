// Package templatefetch is the HTTP client for vendor-supplied stack template
// URLs. Config validation and stack rendering both fetch those URLs, and must
// apply the same policy so a URL that validates also renders.
package templatefetch

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
	fetchTimeout = 30 * time.Second
	dialTimeout  = 10 * time.Second
	maxRedirects = 3

	// MaxBytes caps the body of a fetched template or UI definition.
	MaxBytes = 10 << 20
)

// ErrRefused marks a fetch the policy refused, as opposed to one that failed
// on the network. A refused URL fails the same way on every attempt.
var ErrRefused = errors.New("template fetch refused")

// IsRefused reports whether err is a policy refusal: a disallowed URL,
// redirect, or body size, or a host that is not on the public internet.
func IsRefused(err error) bool {
	return errors.Is(err, ErrRefused) || errors.Is(err, publicnet.ErrNonPublic)
}

type refusedError struct{ msg string }

func (e refusedError) Error() string        { return e.msg }
func (e refusedError) Is(target error) bool { return target == ErrRefused }

func refuse(format string, args ...any) error {
	return refusedError{msg: fmt.Sprintf(format, args...)}
}

// NewClient returns the client used to fetch vendor-supplied template URLs:
// nested stack templates and createUiDefinition files. Those URLs come straight
// from app config, so the fetch must not reach anything but the public
// internet. The client requires https, dials only public addresses (pinning the
// checked address so a DNS rebind cannot swap it), follows a bounded number of
// https-only redirects, and fails reads past MaxBytes.
//
// Every environment, including local dev, fetches templates from remote
// sources; only tests need NewLocalClient.
func NewClient() *http.Client {
	return newClient(false)
}

// NewLocalClient is for tests that serve templates from an httptest server: it
// drops the https and public-network checks. The body cap still applies. It
// must not be used outside tests.
func NewLocalClient() *http.Client {
	return newClient(true)
}

func newClient(allowInsecure bool) *http.Client {
	dialer := &net.Dialer{Timeout: dialTimeout, KeepAlive: 30 * time.Second}
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = nil
	base.DialContext = publicnet.DialContext(dialer)
	if allowInsecure {
		base.DialContext = dialer.DialContext
	}

	return &http.Client{
		Timeout:   fetchTimeout,
		Transport: transport{base: base, allowInsecure: allowInsecure},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > maxRedirects {
				return refuse("template fetch: too many redirects")
			}
			req.Header.Del("Referer")
			return validateURL(req.URL, allowInsecure)
		},
	}
}

type transport struct {
	base          http.RoundTripper
	allowInsecure bool
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := validateURL(req.URL, t.allowInsecure); err != nil {
		return nil, err
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.ContentLength > MaxBytes {
		resp.Body.Close()
		return nil, refuse("template fetch: response exceeds %d bytes", MaxBytes)
	}
	resp.Body = &cappedBody{ReadCloser: resp.Body, remaining: MaxBytes}
	return resp, nil
}

func validateURL(u *url.URL, allowInsecure bool) error {
	if u.Opaque != "" || u.User != nil || u.Hostname() == "" {
		return refuse("template fetch: URL must be an absolute URL without credentials")
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if allowInsecure {
			return nil
		}
		return refuse("template fetch: URL must use https")
	default:
		return refuse("template fetch: unsupported URL scheme %q", u.Scheme)
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
		return 0, refuse("template fetch: response exceeds %d bytes", MaxBytes)
	}
	if int64(len(p)) > b.remaining+1 {
		p = p[:b.remaining+1]
	}
	n, err := b.ReadCloser.Read(p)
	b.remaining -= int64(n)
	if b.remaining < 0 {
		return n, refuse("template fetch: response exceeds %d bytes", MaxBytes)
	}
	return n, err
}
