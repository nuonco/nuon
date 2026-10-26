// Package oidcverify verifies OIDC ID tokens against admin-configured issuers.
//
// The issuer a token is verified against always comes from stored configuration,
// never from the token itself. Discovery and JWKS fetches go through a transport
// that refuses to reach local or internal networks.
package oidcverify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"time"
)

const (
	requestTimeout   = 10 * time.Second
	dialTimeout      = 5 * time.Second
	maxResponseBytes = 1 << 20
	maxRedirects     = 2
)

// nonPublicPrefixes covers ranges net.IP does not classify as private:
// carrier-grade NAT, IETF protocol assignments, benchmarking, and documentation.
var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("2001:db8::/32"),
}

// RequirePublicHost resolves host and rejects it unless every address is on the
// public internet. The dial path checks again, so a record that changes between
// this check and the connection is still caught.
func RequirePublicHost(ctx context.Context, host string) error {
	if ip := net.ParseIP(host); ip != nil {
		if publicIP(ip) {
			return nil
		}
		return fmt.Errorf("%s is not a public address", host)
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("resolving %s: %w", host, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("%s did not resolve", host)
	}
	for _, resolved := range ips {
		if !publicIP(resolved.IP) {
			return fmt.Errorf("%s resolved to a non-public address", host)
		}
	}
	return nil
}

func publicIP(ip net.IP) bool {
	if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()
	for _, prefix := range nonPublicPrefixes {
		if prefix.Contains(addr) {
			return false
		}
	}
	return true
}

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
	base.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		if allowInsecure {
			return dialer.DialContext(ctx, network, net.JoinHostPort(host, port))
		}

		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, errors.New("oidc endpoint did not resolve")
		}
		for _, resolved := range ips {
			if !publicIP(resolved.IP) {
				return nil, fmt.Errorf("oidc endpoint %s resolved to a non-public address", host)
			}
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
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
