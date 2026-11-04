package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRunnerComponentProbes(t *testing.T) {
	assert.Nil(t, runnerComponentProbes(nil))

	probes := runnerComponentProbes(app.ComponentHealthProbes{
		{Type: "http", Name: "api", URL: "https://app.example.com/healthz"},
		{Type: "exec", Command: []string{"check-api", "--fast"}},
		{Type: "EXEC", Command: []string{"/bin/sh", "-c", "true"}},
	})
	require.Len(t, probes, 1)
	assert.Equal(t, RunnerComponentProbe{
		Type: "http",
		Name: "api",
		URL:  "https://app.example.com/healthz",
	}, probes[0])
}

func TestRenderProbe(t *testing.T) {
	stateData := map[string]any{
		"nuon": map[string]any{
			"install": map[string]any{
				"sandbox": map[string]any{
					"outputs": map[string]any{"public_domain": "app.customer.com"},
				},
			},
		},
	}

	t.Run("renders url", func(t *testing.T) {
		got, err := renderProbe(RunnerComponentProbe{
			Type: "http",
			URL:  "https://{{.nuon.install.sandbox.outputs.public_domain}}/healthz",
		}, stateData)
		require.NoError(t, err)
		assert.Equal(t, "https://app.customer.com/healthz", got.URL)
	})

	t.Run("errors on unresolvable target", func(t *testing.T) {
		_, err := renderProbe(RunnerComponentProbe{
			Type: "http",
			URL:  "https://{{.nuon.install.sandbox.outputs.nope}}/healthz",
		}, stateData)
		require.Error(t, err)
	})
}
