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
	require.True(t, requestHostAllowed(hosts, "127.0.0.1:8081"))
	require.True(t, requestHostAllowed(hosts, "localhost:9090"))
	require.False(t, requestHostAllowed(hosts, "evil.example:4321"))
}

func TestAllowedHostsRequiresExplicitHostForWildcard(t *testing.T) {
	_, err := allowedHosts(":8080", &net.TCPAddr{IP: net.IPv6zero, Port: 8080}, nil)
	require.EqualError(t, err, "--allowed-host is required when --addr binds a wildcard address")
}

func TestConfigRequiresRealDeploymentMetadataByDefault(t *testing.T) {
	require.EqualError(t, (config{}).validate(), "real deployment context is required: pass --state and --stack-name, or use --local-state for development")
	require.NoError(t, (config{state: "s3://bucket/install/state", stackName: "install-stack", bundleKey: "install/bundle/app.tar.zst", bundleURI: "s3://bucket/install/bundle/app.tar.zst", runnerImage: "registry/runner:v1", stackOutputsKey: "install/stack-outputs/outputs.json"}).validate())
	require.EqualError(t, (config{state: "s3://bucket/install/state"}).validate(), "--stack-name is required for a real deployment")
	require.EqualError(t, (config{state: "s3://bucket/install/state", stackName: "install-stack"}).validate(), "real deployment bundle uploads require --bundle-key, --bundle-uri, --runner-image, and --stack-outputs-key")
	require.EqualError(t, (config{state: "/tmp/state", stackName: "install-stack"}).validate(), "--state must be an s3:// URI for a real deployment; use --local-state for a local directory")
}

func TestConfigRequiresExplicitLocalMode(t *testing.T) {
	require.NoError(t, (config{localState: "/tmp/state"}).validate())
	require.EqualError(t, (config{state: "s3://bucket/state", localState: "/tmp/state"}).validate(), "--state and --local-state cannot be used together")
	require.EqualError(t, (config{localState: "/tmp/state", stackName: "install-stack"}).validate(), "--stack-name cannot be used with --local-state")
}

func TestStackOutputsRoot(t *testing.T) {
	for state, want := range map[string]string{
		"s3://bucket/deploy-1/state":       "s3://bucket/deploy-1",
		"s3://bucket/deploy-1/state/":      "s3://bucket/deploy-1",
		"s3://bucket/deploy-1/state/retry": "s3://bucket/deploy-1",
		"s3://bucket/state":                "s3://bucket",
		"/var/lib/nuon/deploy/state":       "/var/lib/nuon/deploy",
		"/var/lib/nuon/deploy/state/retry": "/var/lib/nuon/deploy",
		"/state/retry":                     "/",
	} {
		parent, ok := stackOutputsRoot(state)
		require.True(t, ok, state)
		require.Equal(t, want, parent, state)
	}
	for _, state := range []string{"s3://bucket/deploy-1/other", "s3://state", "s3://state/retry", "/tmp/portal-state", "/tmp/stateful/retry"} {
		_, ok := stackOutputsRoot(state)
		require.False(t, ok, state)
	}
}

func TestAllowedHostsAcceptsExplicitDeploymentHost(t *testing.T) {
	hosts, err := allowedHosts(":8080", &net.TCPAddr{IP: net.IPv6zero, Port: 8080}, []string{"portal.internal"})
	require.NoError(t, err)
	require.True(t, hosts["portal.internal"])
	require.True(t, hosts["portal.internal:8080"])
	require.False(t, hosts["evil.example:8080"])
}
