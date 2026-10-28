package publicnet

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPublicIP(t *testing.T) {
	cases := map[string]bool{
		"8.8.8.8":         true,
		"1.1.1.1":         true,
		"127.0.0.1":       false,
		"10.1.2.3":        false,
		"172.16.5.5":      false,
		"192.168.1.1":     false,
		"169.254.169.254": false, // link local, cloud metadata
		"100.64.1.1":      false, // carrier-grade NAT
		"192.0.0.8":       false, // IETF protocol assignments
		"198.18.0.1":      false, // benchmarking
		"0.0.0.0":         false,
		"224.0.0.1":       false, // multicast
		"::1":             false,
		"fe80::1":         false,
		"2001:db8::1":     false, // documentation
		"2606:4700::1":    true,

		"100.100.100.200":        false, // alibaba metadata
		"fd00:ec2::254":          false, // aws metadata
		"fd20:ce::254":           false, // gcp metadata
		"::ffff:169.254.169.254": false, // ipv4-mapped
		"64:ff9b::a9fe:a9fe":     false, // nat64 of 169.254.169.254
		"64:ff9b::a00:102":       false, // nat64 of 10.0.1.2
		"64:ff9b::7f00:1":        false, // nat64 of 127.0.0.1
		"64:ff9b::b9c7:6f85":     true,  // nat64 of 185.199.111.133
		"64:ff9b:1::b9c7:6f85":   false, // local-use nat64, embedding not fixed
		"2002:a9fe:a9fe::1":      false, // 6to4
		"2001:0:a9fe:a9fe::1":    false, // teredo
		"0.1.2.3":                false,
		"240.0.0.1":              false,
		"255.255.255.255":        false,
		"203.0.113.5":            false, // documentation
		"2a00:1450:4001::1":      true,
	}

	for raw, want := range cases {
		require.Equal(t, want, IsPublicIP(net.ParseIP(raw)), raw)
	}
}
