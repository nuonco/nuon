package componenthealth

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func loopbackProber(t *testing.T) *prober {
	t.Helper()

	p := newProberWithControl(nil)
	t.Cleanup(p.close)
	return p
}

func TestProbeAddrBlocked(t *testing.T) {
	blocked := []string{
		"127.0.0.1",
		"127.1.2.3",
		"::1",
		"0.0.0.0",
		"::",
		"169.254.169.254",
		"169.254.170.2",
		"fe80::1",
		"fd00:ec2::254",
		"fd00:ec2::23",
		"fd20:ce::254",
		"64:ff9b:1::a9fe:a9fe",
		"168.63.129.16",
		"100.100.100.200",
		"192.0.0.192",
		"::ffff:169.254.169.254",
		"::ffff:127.0.0.1",
		"64:ff9b::a9fe:a9fe",
		"64:ff9b::7f00:1",
		"224.0.0.1",
		"ff02::1",
	}
	for _, s := range blocked {
		t.Run("blocks_"+s, func(t *testing.T) {
			assert.True(t, probeAddrBlocked(netip.MustParseAddr(s)))
		})
	}

	allowed := []string{
		"10.0.0.5",
		"172.16.4.1",
		"192.168.1.10",
		"100.64.0.10",
		"fd12:3456::1",
		"8.8.8.8",
		"2001:4860:4860::8888",
		"64:ff9b::808:808",
	}
	for _, s := range allowed {
		t.Run("allows_"+s, func(t *testing.T) {
			assert.False(t, probeAddrBlocked(netip.MustParseAddr(s)))
		})
	}
}

func TestGuardProbeDestination(t *testing.T) {
	assert.ErrorIs(t, guardProbeDestination("tcp4", "169.254.169.254:80", nil), errProbeDestinationBlocked)
	assert.ErrorIs(t, guardProbeDestination("tcp6", "[fe80::1%eth0]:80", nil), errProbeDestinationBlocked)
	assert.ErrorIs(t, guardProbeDestination("tcp", "not-an-address", nil), errProbeDestinationBlocked)
	assert.NoError(t, guardProbeDestination("tcp4", "10.0.0.5:5432", nil))
	assert.NoError(t, guardProbeDestination("tcp6", "[2001:4860:4860::8888]:443", nil))
}

func TestProbesRefuseBlockedDestinations(t *testing.T) {
	p := newProber()
	t.Cleanup(p.close)

	t.Run("tcp_metadata", func(t *testing.T) {
		res := p.runTCP(context.Background(), "169.254.169.254:80")
		assert.Equal(t, healthUnhealthy, res.health)
		assert.Equal(t, probeDestinationBlocked, res.message)
	})

	t.Run("tcp_open_loopback_port", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer ln.Close()

		res := p.runTCP(context.Background(), ln.Addr().String())
		assert.Equal(t, healthUnhealthy, res.health)
		assert.Equal(t, probeDestinationBlocked, res.message)
	})

	t.Run("http_loopback_never_receives_the_request", func(t *testing.T) {
		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hits.Add(1)
		}))
		defer srv.Close()

		res := p.runHTTP(context.Background(), srv.URL+"/")
		assert.Equal(t, healthUnhealthy, res.health)
		assert.Equal(t, probeDestinationBlocked, res.message)
		assert.Zero(t, hits.Load())
	})

	t.Run("http_hostname_resolving_to_loopback", func(t *testing.T) {
		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hits.Add(1)
		}))
		defer srv.Close()

		_, port, err := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
		require.NoError(t, err)

		res := p.runHTTP(context.Background(), "http://localhost:"+port+"/")
		assert.Equal(t, healthUnhealthy, res.health)
		assert.Equal(t, probeDestinationBlocked, res.message)
		assert.Zero(t, hits.Load())
	})

	t.Run("http_alternate_loopback_spellings", func(t *testing.T) {
		for _, host := range []string{"127.1", "2130706433", "0x7f000001", "[::ffff:127.0.0.1]"} {
			res := p.runHTTP(context.Background(), "http://"+host+":9091/")
			assert.Equal(t, healthUnhealthy, res.health, host)
			assert.Equal(t, probeDestinationBlocked, res.message, host)
		}
	})

	t.Run("http_runner_health_server", func(t *testing.T) {
		res := p.runHTTP(context.Background(), "http://127.0.0.1:9091/")
		assert.Equal(t, healthUnhealthy, res.health)
		assert.Equal(t, probeDestinationBlocked, res.message)
	})
}

func TestBlockedProbeReportsUnhealthyWithReason(t *testing.T) {
	specs := probeSpecsFor(&models.ServiceRunnerInstallComponent{
		Probes: []*models.ServiceRunnerComponentProbe{
			{Type: "tcp", Name: "imds", URL: "169.254.169.254:80"},
		},
	})
	require.Len(t, specs, 1)

	p := newProber()
	t.Cleanup(p.close)

	row := probeResourceRow(specs[0], p.run(context.Background(), specs[0]))
	assert.Equal(t, healthUnhealthy, row.Health)
	assert.Equal(t, probeDestinationBlocked, row.Message)
	assert.NotContains(t, row.Message, "169.254.169.254")
}

func TestProbeTransportNeverProxies(t *testing.T) {
	p := newProber()
	t.Cleanup(p.close)

	transport, ok := p.client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Nil(t, transport.Proxy)
}
