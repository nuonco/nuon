// Package oidcverify verifies OIDC ID tokens against admin-configured issuers.
//
// The issuer a token is verified against always comes from stored configuration,
// never from the token itself. Discovery and JWKS fetches go through a transport
// that refuses to reach local or internal networks.
package oidcverify

import (
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/nuonco/nuon/pkg/publicnet"
)

const (
	requestTimeout   = 10 * time.Second
	dialTimeout      = 5 * time.Second
	maxResponseBytes = 1 << 20
	maxRedirects     = 2
)

type limitedTransport struct {
	base          http.RoundTripper
	allowInsecure bool
}

func (t limitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.allowInsecure && req.URL.Scheme != "https" {
		return nil, errors.New("oidc requests require https")
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.ContentLength > maxResponseBytes {
		resp.Body.Close()
		return nil, errors.New("oidc response is too large")
	}
	resp.Body = struct {
		io.Reader
		io.Closer
	}{Reader: io.LimitReader(resp.Body, maxResponseBytes), Closer: resp.Body}
	return resp, nil
}

// newHTTPClient returns the client used for discovery and JWKS fetches.
//
// allowInsecure disables the public-network checks so tests and local
// development can point at an issuer on localhost. Redirects stay on https
// either way.
func newHTTPClient(allowInsecure bool) *http.Client {
	dialer := &net.Dialer{Timeout: dialTimeout, KeepAlive: 30 * time.Second}
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = nil
	base.DialContext = publicnet.DialContext(dialer)
	if allowInsecure {
		base.DialContext = dialer.DialContext
	}

	return &http.Client{
		Timeout:   requestTimeout,
		Transport: limitedTransport{base: base, allowInsecure: allowInsecure},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > maxRedirects {
				return errors.New("invalid oidc redirect")
			}
			if !allowInsecure && req.URL.Scheme != "https" {
				return errors.New("invalid oidc redirect")
			}
			return nil
		},
	}
}
