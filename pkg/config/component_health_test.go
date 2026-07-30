package config

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

func TestComponentHealthProbesDecode(t *testing.T) {
	raw := []byte(`
chart_name = "api-chart"
namespace = "default"

[health]
enabled = true

[[health.probes]]
type = "http"
url = "https://{{.nuon.install.sandbox.outputs.public_domain}}/healthz"
# removed key: a config still setting it must decode, not fail sync
interval = "30s"

[[health.probes]]
type = "exec"
name = "custom"
command = ["check-api", "--fast"]
`)

	obj := map[string]any{}
	require.NoError(t, toml.Unmarshal(raw, &obj))

	var out HelmChartComponentConfig
	decCfg := DecoderConfig()
	decCfg.Result = &out
	dec, err := mapstructure.NewDecoder(decCfg)
	require.NoError(t, err)
	require.NoError(t, dec.Decode(obj))

	require.NotNil(t, out.Health)
	require.Len(t, out.Health.Probes, 2)
	require.Equal(t, "http", out.Health.Probes[0].Type)
	require.Equal(t, "https://{{.nuon.install.sandbox.outputs.public_domain}}/healthz", out.Health.Probes[0].URL)
	require.Equal(t, "exec", out.Health.Probes[1].Type)
	require.Equal(t, "custom", out.Health.Probes[1].Name)
	require.Equal(t, []string{"check-api", "--fast"}, out.Health.Probes[1].Command)
	require.NoError(t, out.Validate())
}

func TestComponentHealthConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		probe   ComponentHealthProbeConfig
		wantErr bool
	}{
		{
			name:  "http with url",
			probe: ComponentHealthProbeConfig{Type: "http", URL: "https://app.example.com/healthz"},
		},
		{
			name:  "http with template url",
			probe: ComponentHealthProbeConfig{Type: "http", URL: "https://{{.nuon.install.sandbox.outputs.public_domain}}/healthz"},
		},
		{
			name:    "http without url",
			probe:   ComponentHealthProbeConfig{Type: "http"},
			wantErr: true,
		},
		{
			name:  "tcp with host port",
			probe: ComponentHealthProbeConfig{Type: "tcp", URL: "db.internal:5432"},
		},
		{
			name:    "tcp without url",
			probe:   ComponentHealthProbeConfig{Type: "tcp", URL: "  "},
			wantErr: true,
		},
		{
			name:  "exec with command",
			probe: ComponentHealthProbeConfig{Type: "exec", Command: []string{"check-api", "--fast"}},
		},
		{
			name:    "exec without command",
			probe:   ComponentHealthProbeConfig{Type: "exec"},
			wantErr: true,
		},
		{
			name:    "exec with blank binary",
			probe:   ComponentHealthProbeConfig{Type: "exec", Command: []string{"  "}},
			wantErr: true,
		},
		{
			name:    "unknown type",
			probe:   ComponentHealthProbeConfig{Type: "grpc", URL: "https://app.example.com/"},
			wantErr: true,
		},
		{
			name:    "missing type",
			probe:   ComponentHealthProbeConfig{URL: "https://app.example.com/"},
			wantErr: true,
		},
		{
			name:  "uppercase type",
			probe: ComponentHealthProbeConfig{Type: "HTTP", URL: "https://app.example.com/"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ComponentHealthConfig{Probes: []ComponentHealthProbeConfig{tc.probe}}
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected an error, got none")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	t.Run("nil health block", func(t *testing.T) {
		var cfg *ComponentHealthConfig
		if err := cfg.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
