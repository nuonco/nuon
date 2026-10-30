package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestAssess(t *testing.T) {
	orig := Version
	t.Cleanup(func() { Version = orig })

	policy := func(errorFloor, warnFloor string) *models.ServiceCLIConfig {
		return &models.ServiceCLIConfig{
			MinCliVersion: warnFloor,
			MinCliVersionPolicy: &models.ServiceMinCLIVersionPolicy{
				Error: errorFloor,
				Warn:  warnFloor,
			},
		}
	}

	tests := []struct {
		name    string
		version string
		cfg     *models.ServiceCLIConfig
		warn    string
		block   bool
	}{
		{
			name:    "development build",
			version: "development",
			cfg:     policy("0.19.1102", "0.19.1400"),
		},
		{
			name:    "commit sha",
			version: "abc1234",
			cfg:     policy("0.19.1102", "0.19.1400"),
		},
		{
			name:    "below error",
			version: "0.19.1000",
			cfg:     policy("0.19.1102", "0.19.1400"),
			block:   true,
		},
		{
			name:    "between error and warn",
			version: "0.19.1200",
			cfg:     policy("0.19.1102", "0.19.1400"),
			warn:    "Your CLI (0.19.1200) is older than the minimum 0.19.1400 required by this control plane. see https://docs.nuon.co/cli to update.",
		},
		{
			name:    "at warn",
			version: "0.19.1400",
			cfg:     policy("0.19.1102", "0.19.1400"),
		},
		{
			name:    "empty warn",
			version: "0.19.1001",
			cfg:     policy("0.19.1000", ""),
		},
		{
			name:    "warn below error is ignored once error is met",
			version: "0.19.1200",
			cfg:     policy("0.19.1102", "0.19.1000"),
		},
		{
			name:    "legacy string warns when policy is absent",
			version: "0.19.1200",
			cfg:     &models.ServiceCLIConfig{MinCliVersion: "0.19.1400"},
			warn:    "Your CLI (0.19.1200) is older than the minimum 0.19.1400 required by this control plane. see https://docs.nuon.co/cli to update.",
		},
		{
			name:    "policy with empty warn ignores the legacy string",
			version: "0.19.1200",
			cfg: &models.ServiceCLIConfig{
				MinCliVersion:       "9.0.0",
				MinCliVersionPolicy: &models.ServiceMinCLIVersionPolicy{Error: "0.19.1000"},
			},
		},
		{
			name:    "empty config",
			version: "0.19.1200",
			cfg:     &models.ServiceCLIConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			warn, block := Assess(tt.cfg)
			assert.Equal(t, tt.warn, warn)
			if tt.block {
				require.Error(t, block)
				assert.Equal(t, "Your CLI (0.19.1000) is older than 0.19.1102, the oldest version this control plane accepts. see https://docs.nuon.co/cli to update.", block.Error())
				return
			}
			assert.NoError(t, block)
		})
	}
}
