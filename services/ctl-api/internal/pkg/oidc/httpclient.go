package oidc

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

const (
	defaultHTTPTimeout  = 10 * time.Second
	defaultMaxBodyBytes = int64(1 << 20)
)

var (
	ErrResponseTooLarge = errors.New("oidc: response body exceeds maximum size")
	errNonPublicAddress = errors.New("oidc: refusing to connect to non-public address")
	errRedirect         = errors.New("oidc: redirects are not followed")
)

var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001::/32"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("fc00::/7"),
}

type HTTPClientOptions struct {
	Timeout               time.Duration
	MaxBodyBytes          int64
	AllowPrivateAddresses bool
}

func NewHTTPClient(opts HTTPClientOptions) *http.Client {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}
	maxBody := opts.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = defaultMaxBodyBytes
	}

	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}
	if !opts.AllowPrivateAddresses {
		dialer.Control = guardControl
	}

	var transport *http.Transport
	if base, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = base.Clone()
	} else {
		transport = &http.Transport{}
	}
	transport.Proxy = nil
	transport.DialContext = dialer.DialContext
	transport.TLSHandshakeTimeout = min(timeout, 10*time.Second)
	transport.ResponseHeaderTimeout = timeout
	transport.ExpectContinueTimeout = time.Second

	return &http.Client{
		Timeout:   timeout,
		Transport: &limitedTransport{base: transport, max: maxBody},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errRedirect
		},
	}
}

func guardControl(_, address string, _ syscall.RawConn) error {
	addrPort, err := netip.ParseAddrPort(address)
	if err != nil || isBlockedAddr(addrPort.Addr()) {
		return errNonPublicAddress
	}
	return nil
}

func isBlockedAddr(addr netip.Addr) bool {
	addr = addr.Unmap().WithZone("")
	if !addr.IsValid() {
		return true
	}
	if addr.IsLoopback() ||
		addr.IsPrivate() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsInterfaceLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified() {
		return true
	}
	for _, p := range blockedPrefixes {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

type limitedTransport struct {
	base http.RoundTripper
	max  int64
}

func (t *limitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp.Body = &limitedBody{rc: resp.Body, remaining: t.max}
	return resp, nil
}

type limitedBody struct {
	rc        io.ReadCloser
	remaining int64
}

func (b *limitedBody) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if b.remaining < 0 {
		return 0, ErrResponseTooLarge
	}
	if int64(len(p)) > b.remaining+1 {
		p = p[:b.remaining+1]
	}
	n, err := b.rc.Read(p)
	if int64(n) > b.remaining {
		n = int(b.remaining)
		b.remaining = -1
		return n, ErrResponseTooLarge
	}
	b.remaining -= int64(n)
	return n, err
}

func (b *limitedBody) Close() error {
	return b.rc.Close()
}
