package componenthealth

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"syscall"

	"github.com/nuonco/nuon/pkg/publicnet"
)

const probeDestinationBlocked = "probe destination is not allowed: loopback, link-local and cloud metadata addresses cannot be probed"

var errProbeDestinationBlocked = errors.New(probeDestinationBlocked)

var nat64Prefix = netip.MustParsePrefix("64:ff9b::/96")

// Cloud metadata addresses come from publicnet.IsMetadataAddr, so this list and
// publicnet's cannot drift.
var blockedProbePrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
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
	if publicnet.IsMetadataAddr(addr) {
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
