package config

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

func TestCustomerManagedRuntimeParsesPlatformSpecificArtifacts(t *testing.T) {
	var cfg AppConfig
	require.NoError(t, toml.Unmarshal([]byte(`
[customer_managed]
runner_image_url = "registry.example.com/runner"
runner_image_tag = "v1"

[[customer_managed.platforms]]
target = "linux/amd64"
portal_binary_url = "https://artifacts.example.com/portal-amd64"
runner_binary_url = "https://artifacts.example.com/runner-amd64"
`), &cfg))
	require.NotNil(t, cfg.CustomerManaged)
	require.Equal(t, "registry.example.com/runner", cfg.CustomerManaged.RunnerImageURL)
	require.Equal(t, "v1", cfg.CustomerManaged.RunnerImageTag)
	require.Equal(t, []CustomerManagedPlatform{{
		Target:          "linux/amd64",
		PortalBinaryURL: "https://artifacts.example.com/portal-amd64",
		RunnerBinaryURL: "https://artifacts.example.com/runner-amd64",
	}}, cfg.CustomerManaged.Platforms)
}
