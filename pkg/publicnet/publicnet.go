// Package publicnet restricts outbound connections to the public internet, for
// fetches whose destination comes from untrusted input.
package publicnet

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
)

// globalUnicastV6 is the only IPv6 range treated as public, apart from
// well-known NAT64 addresses whose embedded IPv4 address is public.
var globalUnicastV6 = netip.MustParsePrefix("2000::/3")

// nat64WellKnown is where DNS64 resolvers on IPv6-only networks synthesize
// addresses for IPv4-only hosts. The low 32 bits are the IPv4 destination.
var nat64WellKnown = netip.MustParsePrefix("64:ff9b::/96")

// nonPublicPrefixes covers ranges net.IP does not classify as private.
var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),       // "this network"
	netip.MustParsePrefix("100.64.0.0/10"),   // carrier-grade NAT, Alibaba metadata
	netip.MustParsePrefix("192.0.0.0/24"),    // IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),    // documentation
	netip.MustParsePrefix("198.18.0.0/15"),   // benchmarking
	netip.MustParsePrefix("198.51.100.0/24"), // documentation
	netip.MustParsePrefix("203.0.113.0/24"),  // documentation
	netip.MustParsePrefix("240.0.0.0/4"),     // reserved, broadcast
	netip.MustParsePrefix("2001::/32"),       // Teredo, embeds an IPv4 address
	netip.MustParsePrefix("2001:db8::/32"),   // documentation
	netip.MustParsePrefix("2002::/16"),       // 6to4, embeds an IPv4 address
}

func IsPublicIP(ip net.IP) bool {
	if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()
	if nat64WellKnown.Contains(addr) {
		embedded := addr.As16()
		return IsPublicIP(net.IP(embedded[12:]))
	}
	if addr.Is6() && !globalUnicastV6.Contains(addr) {
		return false
	}
	for _, prefix := range nonPublicPrefixes {
		if prefix.Contains(addr) {
			return false
		}
	}
	return true
}

// ResolvePublic resolves host and returns its addresses, failing unless every
// address is on the public internet.
func ResolvePublic(ctx context.Context, host string) ([]net.IPAddr, error) {
	if ip := net.ParseIP(host); ip != nil {
		if IsPublicIP(ip) {
			return []net.IPAddr{{IP: ip}}, nil
		}
		return nil, fmt.Errorf("%s is a non-public address", host)
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolving %s: %w", host, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("%s did not resolve", host)
	}
	for _, resolved := range ips {
		if !IsPublicIP(resolved.IP) {
			return nil, fmt.Errorf("%s resolved to a non-public address", host)
		}
	}
	return ips, nil
}

// RequirePublicHost rejects host unless every address it resolves to is on the
// public internet. Callers that then connect should also dial through
// DialContext, so a record that changes between the check and the connection
// is still caught.
func RequirePublicHost(ctx context.Context, host string) error {
	_, err := ResolvePublic(ctx, host)
	return err
}

// DialContext wraps dialer so it only connects to public addresses, pinning the
// connection to the address that was checked.
func DialContext(dialer *net.Dialer) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := ResolvePublic(ctx, host)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, errors.New("no addresses to dial")
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
	}
}
