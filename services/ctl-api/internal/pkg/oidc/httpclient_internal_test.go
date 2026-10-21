package oidc

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressClassification(t *testing.T) {
	tests := map[string]struct {
		ip      string
		blocked bool
	}{
		"ipv4 loopback":      {ip: "127.0.0.1", blocked: true},
		"ipv6 loopback":      {ip: "::1", blocked: true},
		"rfc1918 10":         {ip: "10.0.0.1", blocked: true},
		"rfc1918 172":        {ip: "172.16.0.1", blocked: true},
		"rfc1918 192":        {ip: "192.168.1.1", blocked: true},
		"metadata":           {ip: "169.254.169.254", blocked: true},
		"ipv6 link local":    {ip: "fe80::1", blocked: true},
		"ipv6 ula":           {ip: "fc00::1", blocked: true},
		"unspecified v4":     {ip: "0.0.0.0", blocked: true},
		"unspecified v6":     {ip: "::", blocked: true},
		"cgnat":              {ip: "100.64.0.1", blocked: true},
		"mapped loopback":    {ip: "::ffff:127.0.0.1", blocked: true},
		"mapped private":     {ip: "::ffff:10.0.0.1", blocked: true},
		"multicast v4":       {ip: "224.0.0.1", blocked: true},
		"multicast v6":       {ip: "ff02::1", blocked: true},
		"this network":       {ip: "0.1.2.3", blocked: true},
		"reserved v4":        {ip: "240.0.0.1", blocked: true},
		"broadcast":          {ip: "255.255.255.255", blocked: true},
		"documentation v4":   {ip: "203.0.113.5", blocked: true},
		"nat64 loopback":     {ip: "64:ff9b::7f00:1", blocked: true},
		"6to4":               {ip: "2002:7f00:1::1", blocked: true},
		"teredo":             {ip: "2001::1", blocked: true},
		"public v4":          {ip: "8.8.8.8", blocked: false},
		"public v6":          {ip: "2606:4700:4700::1111", blocked: false},
		"mapped public":      {ip: "::ffff:8.8.8.8", blocked: false},
		"just outside cgnat": {ip: "100.128.0.1", blocked: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.blocked, isBlockedAddr(netip.MustParseAddr(tt.ip)))
		})
	}
}
