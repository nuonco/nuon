package componenthealth

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestNewProbeSpec(t *testing.T) {
	tests := []struct {
		name       string
		kind       string
		target     string
		wantOK     bool
		wantKind   string
		wantTarget string
	}{
		{
			name:       "http",
			kind:       "http",
			target:     "http://api.internal:8080/healthz",
			wantOK:     true,
			wantKind:   probeKindHTTP,
			wantTarget: "http://api.internal:8080/healthz",
		},
		{
			name:       "https_uppercase_kind",
			kind:       "HTTP",
			target:     "  https://app.example.com/healthz  ",
			wantOK:     true,
			wantKind:   probeKindHTTP,
			wantTarget: "https://app.example.com/healthz",
		},
		{
			name:   "http_without_scheme",
			kind:   "http",
			target: "app.example.com/healthz",
		},
		{
			name:   "http_unsupported_scheme",
			kind:   "http",
			target: "ftp://app.example.com/",
		},
		{
			name:   "empty_target",
			kind:   "http",
			target: "   ",
		},
		{
			name:   "unknown_kind",
			kind:   "grpc",
			target: "https://app.example.com/",
		},
		{
			name:       "tcp_host_port",
			kind:       "tcp",
			target:     "db.internal:5432",
			wantOK:     true,
			wantKind:   probeKindTCP,
			wantTarget: "db.internal:5432",
		},
		{
			name:       "tcp_url_default_port",
			kind:       "tcp",
			target:     "https://app.example.com/healthz",
			wantOK:     true,
			wantKind:   probeKindTCP,
			wantTarget: "app.example.com:443",
		},
		{
			name:   "tcp_without_port",
			kind:   "tcp",
			target: "db.internal",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := newProbeSpec(tc.kind, tc.target)
			if !tc.wantOK {
				assert.False(t, ok)
				assert.Empty(t, spec.target)
				return
			}
			require.True(t, ok)
			assert.Equal(t, tc.wantKind, spec.kind)
			assert.Equal(t, tc.wantTarget, spec.target)
			if tc.wantKind == probeKindTCP {
				assert.Equal(t, tc.wantTarget, spec.dialAddr)
			}
		})
	}
}

func TestRunHTTPProbe(t *testing.T) {
	client := newProbeHTTPClient()
	defer client.CloseIdleConnections()

	t.Run("2xx_is_healthy", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, probeUserAgent, r.Header.Get("User-Agent"))
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		res := runHTTPProbe(context.Background(), client, srv.URL+"/healthz")
		assert.Equal(t, healthHealthy, res.health)
		assert.Empty(t, res.message)
		assert.Equal(t, http.StatusNoContent, res.statusCode)
	})

	t.Run("3xx_is_healthy_and_not_followed", func(t *testing.T) {
		var followed atomic.Int32
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			followed.Add(1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer target.Close()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, target.URL, http.StatusFound)
		}))
		defer srv.Close()

		res := runHTTPProbe(context.Background(), client, srv.URL+"/healthz")
		assert.Equal(t, healthHealthy, res.health)
		assert.Equal(t, http.StatusFound, res.statusCode)
		assert.Zero(t, followed.Load())
	})

	t.Run("4xx_is_unhealthy_with_status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		res := runHTTPProbe(context.Background(), client, srv.URL+"/healthz")
		assert.Equal(t, healthUnhealthy, res.health)
		assert.Equal(t, http.StatusNotFound, res.statusCode)
		assert.Contains(t, res.message, "404")
		assert.Contains(t, res.message, srv.URL)
	})

	t.Run("5xx_is_unhealthy", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer srv.Close()

		res := runHTTPProbe(context.Background(), client, srv.URL+"/")
		assert.Equal(t, healthUnhealthy, res.health)
		assert.Contains(t, res.message, "502")
	})

	t.Run("timeout_is_unhealthy", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(2 * time.Second):
			}
		}))
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		res := runHTTPProbe(ctx, client, srv.URL+"/healthz")
		assert.Equal(t, healthUnhealthy, res.health)
		assert.Contains(t, strings.ToLower(res.message), "deadline")
		assert.Zero(t, res.statusCode)
	})

	t.Run("transport_error_is_unhealthy", func(t *testing.T) {
		res := runHTTPProbe(context.Background(), client, "http://"+closedAddr(t)+"/healthz")
		assert.Equal(t, healthUnhealthy, res.health)
		assert.NotEmpty(t, res.message)
	})
}

func TestRunTCPProbe(t *testing.T) {
	t.Run("open_port_is_healthy", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer ln.Close()

		res := runTCPProbe(context.Background(), ln.Addr().String())
		assert.Equal(t, healthHealthy, res.health)
		assert.Empty(t, res.message)
	})

	t.Run("refused_is_unhealthy", func(t *testing.T) {
		res := runTCPProbe(context.Background(), closedAddr(t))
		assert.Equal(t, healthUnhealthy, res.health)
		assert.NotEmpty(t, res.message)
	})
}

func TestProbeSpecsFor(t *testing.T) {
	c := &models.ServiceRunnerInstallComponent{
		Probes: []*models.ServiceRunnerComponentProbe{
			nil,
			{Type: "http", URL: "https://app.example.com/healthz", Name: "api"},
			{Type: "tcp", URL: "db.internal:5432"},
			{Type: "EXEC", Command: []string{"check-api", "--fast"}},
			{Type: "http", URL: "not-a-url"},
			{Type: "exec"},
			{Type: "grpc", URL: "https://app.example.com/"},
		},
	}

	specs := probeSpecsFor(c)
	require.Len(t, specs, 6)

	for _, i := range []int{0, 1} {
		assert.Empty(t, specs[i].unresolved, "spec %d should be runnable", i)
	}
	for _, i := range []int{2, 3, 4, 5} {
		assert.NotEmpty(t, specs[i].unresolved, "spec %d should report why it cannot run", i)
	}

	assert.Equal(t, probeKindHTTP, specs[0].kind)
	assert.Equal(t, "https://app.example.com/healthz", specs[0].target)
	assert.Equal(t, "api", specs[0].displayName())

	assert.Equal(t, probeKindTCP, specs[1].kind)
	assert.Equal(t, "db.internal:5432", specs[1].displayName())

	assert.Equal(t, probeKindExec, specs[2].kind)
	assert.Contains(t, specs[2].unresolved, "not supported")
	assert.Equal(t, "check-api", specs[2].displayName())
	res := runProbe(context.Background(), nil, specs[2])
	assert.Equal(t, healthUnknown, res.health)

	assert.Nil(t, probeSpecsFor(nil))
	assert.Nil(t, probeSpecsFor(&models.ServiceRunnerInstallComponent{}))
}

func TestProbeResourceRow(t *testing.T) {
	t.Run("healthy_http", func(t *testing.T) {
		spec, ok := newProbeSpec("http", "https://app.example.com/healthz")
		require.True(t, ok)

		row := probeResourceRow(spec, probeResult{health: healthHealthy, statusCode: 200, latency: 12 * time.Millisecond})
		assert.Equal(t, providerProbe, row.Provider)
		assert.Equal(t, resourceKindHTTPProbe, row.Kind)
		assert.Equal(t, "https://app.example.com/healthz", row.Name)
		assert.Equal(t, healthHealthy, row.Health)
		assert.Empty(t, row.Namespace)
		assert.Empty(t, row.Message)

		details := map[string]any{}
		require.NoError(t, json.Unmarshal([]byte(row.Details), &details))
		assert.NotContains(t, details, "diagnosis")
		probe, ok := details["probe"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "http", probe["type"])
		assert.Equal(t, "https://app.example.com/healthz", probe["target"])
		assert.EqualValues(t, 200, probe["status_code"])
		assert.EqualValues(t, 12, probe["latency_ms"])
	})

	t.Run("unhealthy_tcp_carries_diagnosis", func(t *testing.T) {
		spec, ok := newProbeSpec("tcp", "db.internal:5432")
		require.True(t, ok)

		row := probeResourceRow(spec, probeResult{health: healthUnhealthy, message: "connection refused"})
		assert.Equal(t, resourceKindTCPProbe, row.Kind)
		assert.Equal(t, "db.internal:5432", row.Name)
		assert.Equal(t, healthUnhealthy, row.Health)
		assert.Equal(t, "connection refused", row.Message)

		details := map[string]any{}
		require.NoError(t, json.Unmarshal([]byte(row.Details), &details))
		diagnosis, ok := details["diagnosis"].(map[string]any)
		require.True(t, ok)
		probe, ok := diagnosis["probe"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "tcp", probe["type"])
		assert.Equal(t, "db.internal:5432", probe["target"])
	})

	t.Run("named_probe_uses_config_name", func(t *testing.T) {
		spec, ok := newProbeSpec("http", "https://app.example.com/healthz")
		require.True(t, ok)
		spec.name = "api-healthz"

		row := probeResourceRow(spec, probeResult{health: healthHealthy, statusCode: 200})
		assert.Equal(t, "api-healthz", row.Name)
	})
}

// closedAddr returns a loopback address with nothing listening on it.
func closedAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())
	return addr
}

// A declared probe that cannot run must still report — as unknown, with the
// reason — rather than silently disappearing, or the vendor thinks it ran.
func TestUnrunnableProbesReportUnknownRatherThanVanish(t *testing.T) {
	tests := []struct {
		name       string
		probe      *models.ServiceRunnerComponentProbe
		wantReason string
		wantKind   string
		wantName   string
	}{
		{
			name:       "unresolved template in url",
			probe:      &models.ServiceRunnerComponentProbe{Type: "http", Name: "public-endpoint", URL: "https://whoami.{{.nuon.install.sandbox.outputs.nuon_dns.public_domain.name}}/health"},
			wantReason: "could not be resolved from install state",
			wantKind:   resourceKindHTTPProbe,
			wantName:   "public-endpoint",
		},
		{
			name:       "exec probe with a command",
			probe:      &models.ServiceRunnerComponentProbe{Type: "exec", Name: "check", Command: []string{"/bin/sh", "-c", "true"}},
			wantReason: "not supported",
			wantKind:   resourceKindExecProbe,
			wantName:   "check",
		},
		{
			name:       "http probe with no url",
			probe:      &models.ServiceRunnerComponentProbe{Type: "http", Name: "no-url"},
			wantReason: "no url to check",
			wantKind:   resourceKindHTTPProbe,
			wantName:   "no-url",
		},
		{
			name:       "exec probe with no command",
			probe:      &models.ServiceRunnerComponentProbe{Type: "exec", Name: "no-cmd"},
			wantReason: "not supported",
			wantKind:   resourceKindExecProbe,
			wantName:   "no-cmd",
		},
		{
			name:       "unsupported type",
			probe:      &models.ServiceRunnerComponentProbe{Type: "carrier-pigeon", Name: "odd", URL: "https://x.test/health"},
			wantReason: "not supported",
			wantKind:   resourceKindHTTPProbe,
			wantName:   "odd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs := probeSpecsFor(&models.ServiceRunnerInstallComponent{
				Probes: []*models.ServiceRunnerComponentProbe{tt.probe},
			})
			require.Len(t, specs, 1, "the probe must still be reported, not dropped")
			require.NotEmpty(t, specs[0].unresolved)
			require.Contains(t, specs[0].unresolved, tt.wantReason)

			res := runProbe(context.Background(), nil, specs[0])
			require.Equal(t, healthUnknown, res.health, "an unrunnable probe is unknown, never healthy and never a failure")
			require.Contains(t, res.message, tt.wantReason)

			row := probeResourceRow(specs[0], res)
			require.Equal(t, providerProbe, row.Provider)
			require.Equal(t, tt.wantKind, row.Kind)
			require.Equal(t, tt.wantName, row.Name)
			require.Equal(t, healthUnknown, row.Health)
		})
	}
}

func TestValidProbeStillRuns(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	specs := probeSpecsFor(&models.ServiceRunnerInstallComponent{
		Probes: []*models.ServiceRunnerComponentProbe{
			{Type: "tcp", Name: "ok", URL: ln.Addr().String()},
		},
	})
	require.Len(t, specs, 1)
	require.Empty(t, specs[0].unresolved)
	require.Equal(t, healthHealthy, runProbe(context.Background(), nil, specs[0]).health)
}
