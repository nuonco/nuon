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

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewExecProbeSpec(t *testing.T) {
	tests := []struct {
		name    string
		command []string
		wantOK  bool
	}{
		{name: "argv", command: []string{"true"}, wantOK: true},
		{name: "argv_with_args", command: []string{"ls", "-la"}, wantOK: true},
		{name: "trims_binary", command: []string{"  true  "}, wantOK: true},
		{name: "empty", command: nil},
		{name: "blank_binary", command: []string{"   ", "-la"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := newExecProbeSpec(tc.command)
			if !tc.wantOK {
				assert.False(t, ok)
				assert.Empty(t, spec.command)
				return
			}
			require.True(t, ok)
			assert.Equal(t, probeKindExec, spec.kind)
			assert.Equal(t, resourceKindExecProbe, spec.resourceKind())
			assert.Equal(t, strings.TrimSpace(tc.command[0]), spec.command[0])
			assert.Empty(t, spec.target)
		})
	}

	t.Run("does_not_alias_config_argv", func(t *testing.T) {
		command := []string{"true", "--flag"}
		spec, ok := newExecProbeSpec(command)
		require.True(t, ok)

		spec.command[1] = "--mutated"
		assert.Equal(t, "--flag", command[1])
	})
}

func TestRunExecProbe(t *testing.T) {
	tests := []struct {
		name         string
		command      []string
		ctxTimeout   time.Duration
		wantHealth   string
		wantExitCode *int
		wantMessage  []string
	}{
		{
			name:       "exit_zero_is_healthy",
			command:    []string{"true"},
			wantHealth: healthHealthy,
		},
		{
			name:         "exit_non_zero_is_unhealthy",
			command:      []string{"false"},
			wantHealth:   healthUnhealthy,
			wantExitCode: ptr(1),
			wantMessage:  []string{"exit code 1"},
		},
		{
			name:        "exit_non_zero_carries_output_tail",
			command:     []string{"ls", "/nuon-component-health-does-not-exist"},
			wantHealth:  healthUnhealthy,
			wantMessage: []string{"exit code", "/nuon-component-health-does-not-exist"},
		},
		{
			name:        "timeout_is_unhealthy",
			command:     []string{"sleep", "30"},
			ctxTimeout:  50 * time.Millisecond,
			wantHealth:  healthUnhealthy,
			wantMessage: []string{"timed out"},
		},
		{
			name:        "missing_binary_is_unhealthy",
			command:     []string{"nuon-component-health-missing-binary"},
			wantHealth:  healthUnhealthy,
			wantMessage: []string{"nuon-component-health-missing-binary"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.ctxTimeout)
				defer cancel()
			}

			res := runExecProbe(ctx, tc.command)
			assert.Equal(t, tc.wantHealth, res.health)
			if tc.wantHealth == healthHealthy {
				assert.Empty(t, res.message)
				assert.Nil(t, res.exitCode)
				return
			}
			if tc.wantExitCode != nil {
				require.NotNil(t, res.exitCode)
				assert.Equal(t, *tc.wantExitCode, *res.exitCode)
			}
			for _, want := range tc.wantMessage {
				assert.Contains(t, res.message, want)
			}
		})
	}

	t.Run("output_tail_is_bounded", func(t *testing.T) {
		res := runExecProbe(context.Background(), []string{"awk", "BEGIN { for (i = 0; i < 5000; i++) print \"noisy-output-line\"; exit 3 }"})
		assert.Equal(t, healthUnhealthy, res.health)
		require.NotNil(t, res.exitCode)
		assert.Equal(t, 3, *res.exitCode)
		assert.LessOrEqual(t, len(res.message), probeOutputLimit+len("exit code 3: "))
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
	// Every declared probe is reported: the three runnable ones plus the three
	// that cannot run, which report unknown with a reason rather than vanishing.
	require.Len(t, specs, 6)

	for _, i := range []int{0, 1, 2} {
		assert.Empty(t, specs[i].unresolved, "spec %d should be runnable", i)
	}
	for _, i := range []int{3, 4, 5} {
		assert.NotEmpty(t, specs[i].unresolved, "spec %d should report why it cannot run", i)
	}

	assert.Equal(t, probeKindHTTP, specs[0].kind)
	assert.Equal(t, "https://app.example.com/healthz", specs[0].target)
	assert.Equal(t, "api", specs[0].displayName())

	assert.Equal(t, probeKindTCP, specs[1].kind)
	assert.Equal(t, "db.internal:5432", specs[1].displayName())

	assert.Equal(t, probeKindExec, specs[2].kind)
	assert.Equal(t, []string{"check-api", "--fast"}, specs[2].command)
	assert.Equal(t, "check-api", specs[2].displayName())

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

	t.Run("unhealthy_exec_carries_diagnosis", func(t *testing.T) {
		spec, ok := newExecProbeSpec([]string{"check-api", "--fast"})
		require.True(t, ok)

		row := probeResourceRow(spec, probeResult{
			health:   healthUnhealthy,
			message:  "exit code 2: boom",
			exitCode: ptr(2),
		})
		assert.Equal(t, providerProbe, row.Provider)
		assert.Equal(t, resourceKindExecProbe, row.Kind)
		assert.Equal(t, "check-api", row.Name)

		details := map[string]any{}
		require.NoError(t, json.Unmarshal([]byte(row.Details), &details))
		diagnosis, ok := details["diagnosis"].(map[string]any)
		require.True(t, ok)
		probe, ok := diagnosis["probe"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "exec", probe["type"])
		assert.NotContains(t, probe, "target")
		assert.EqualValues(t, 2, probe["exit_code"])
		assert.Equal(t, []any{"check-api", "--fast"}, probe["command"])
	})

	t.Run("named_probe_uses_config_name", func(t *testing.T) {
		spec, ok := newProbeSpec("http", "https://app.example.com/healthz")
		require.True(t, ok)
		spec.name = "api-healthz"

		row := probeResourceRow(spec, probeResult{health: healthHealthy, statusCode: 200})
		assert.Equal(t, "api-healthz", row.Name)
	})
}

func ptr[T any](v T) *T {
	return &v
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

// A vendor's health check must not inherit the runner's control-plane token or
// cloud credential env vars.
func TestRunExecProbeEnvIsolation(t *testing.T) {
	t.Setenv("NUON_TEST_SECRET_SENTINEL", "leaked")

	res := runExecProbe(context.Background(), []string{"sh", "-c", `test -z "$NUON_TEST_SECRET_SENTINEL"`})
	require.Equal(t, healthHealthy, res.health, "runner env leaked into the probe: %s", res.message)

	res = runExecProbe(context.Background(), []string{"sh", "-c", `test -n "$PATH" && test -n "$HOME"`})
	require.Equal(t, healthHealthy, res.health, "probe env is missing PATH/HOME: %s", res.message)
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
			name:       "unresolved template in exec argv",
			probe:      &models.ServiceRunnerComponentProbe{Type: "exec", Name: "check", Command: []string{"/bin/check", "--host={{.nuon.install.sandbox.outputs.host}}"}},
			wantReason: "could not be resolved from install state",
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
			wantReason: "no command to run",
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

// A valid probe must be unaffected by the unresolved path.
func TestValidProbeStillRuns(t *testing.T) {
	specs := probeSpecsFor(&models.ServiceRunnerInstallComponent{
		Probes: []*models.ServiceRunnerComponentProbe{
			{Type: "exec", Name: "ok", Command: []string{"true"}},
		},
	})
	require.Len(t, specs, 1)
	require.Empty(t, specs[0].unresolved)
	require.Equal(t, healthHealthy, runProbe(context.Background(), nil, specs[0]).health)
}
