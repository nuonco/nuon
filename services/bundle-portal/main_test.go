package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllowedHostsForLoopback(t *testing.T) {
	hosts, err := allowedHosts("127.0.0.1:0", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 4321}, nil)
	require.NoError(t, err)
	require.True(t, hosts["localhost:4321"])
	require.True(t, hosts["127.0.0.1:4321"])
	require.False(t, hosts["evil.example"])
}

func TestAllowedHostsRequiresExplicitHostForWildcard(t *testing.T) {
	_, err := allowedHosts(":8080", &net.TCPAddr{IP: net.IPv6zero, Port: 8080}, nil)
	require.EqualError(t, err, "--allowed-host is required when --addr binds a wildcard address")
}

func TestAllowedHostsAcceptsExplicitDeploymentHost(t *testing.T) {
	hosts, err := allowedHosts(":8080", &net.TCPAddr{IP: net.IPv6zero, Port: 8080}, []string{"portal.internal"})
	require.NoError(t, err)
	require.True(t, hosts["portal.internal"])
	require.True(t, hosts["portal.internal:8080"])
	require.False(t, hosts["evil.example:8080"])
}
