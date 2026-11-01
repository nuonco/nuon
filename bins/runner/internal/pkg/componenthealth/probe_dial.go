package componenthealth

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"syscall"
)

const probeDestinationBlocked = "probe destination is not allowed: loopback, link-local and cloud metadata addresses cannot be probed"

var errProbeDestinationBlocked = errors.New(probeDestinationBlocked)

var nat64Prefix = netip.MustParsePrefix("64:ff9b::/96")

var blockedProbePrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.100.100.200/32"),
	netip.MustParsePrefix("168.63.129.16/32"),
	netip.MustParsePrefix("192.0.0.192/32"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("fd00:ec2::23/128"),
	netip.MustParsePrefix("fd00:ec2::254/128"),
	netip.MustParsePrefix("fd20:ce::254/128"),
}

func probeAddrBlocked(addr netip.Addr) bool {
	addr = addr.Unmap()
	if nat64Prefix.Contains(addr) {
		b := addr.As16()
		addr = netip.AddrFrom4([4]byte(b[12:]))
	}

	if addr.IsLoopback() || addr.IsUnspecified() || addr.IsLinkLocalUnicast() || addr.IsMulticast() {
		return true
	}
	for _, prefix := range blockedProbePrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func guardProbeDestination(_, address string, _ syscall.RawConn) error {
	addrPort, err := netip.ParseAddrPort(address)
	if err != nil {
		return errProbeDestinationBlocked
	}
	if probeAddrBlocked(addrPort.Addr()) {
		return errProbeDestinationBlocked
	}
	return nil
}

type prober struct {
	dialer *net.Dialer
	client *http.Client
}

func newProber() *prober {
	return newProberWithControl(guardProbeDestination)
}

// newProberWithControl owns its transport because the runner mutates the default one.
func newProberWithControl(control func(network, address string, c syscall.RawConn) error) *prober {
	dialer := &net.Dialer{Timeout: probeTimeout, Control: control}
	return &prober{
		dialer: dialer,
		client: &http.Client{
			Timeout: probeTimeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				Proxy:                 nil,
				DialContext:           dialer.DialContext,
				TLSHandshakeTimeout:   probeTimeout,
				ResponseHeaderTimeout: probeTimeout,
				DisableKeepAlives:     true,
			},
		},
	}
}

func (p *prober) close() {
	p.client.CloseIdleConnections()
}

func blockedProbeResult() probeResult {
	return probeResult{health: healthUnhealthy, message: probeDestinationBlocked}
}
