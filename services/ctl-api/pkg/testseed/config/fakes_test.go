package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetMinimalAppConfig verifies that GetMinimalAppConfig returns a valid config.
func TestGetMinimalAppConfig(t *testing.T) {
	cfg := GetMinimalAppConfig()

	require.NotNil(t, cfg, "GetMinimalAppConfig should return a config")
	assert.Equal(t, "1", cfg.Version, "Version should be '1'")
	assert.NotEmpty(t, cfg.DisplayName, "DisplayName should be set")
	assert.NotEmpty(t, cfg.Description, "Description should be set")
	assert.NotNil(t, cfg.Sandbox, "Sandbox should be set")
	assert.NotNil(t, cfg.Runner, "Runner should be set")
	assert.NotNil(t, cfg.Components, "Components should be initialized")
	assert.NotNil(t, cfg.Actions, "Actions should be initialized")
	assert.Empty(t, cfg.Components, "Components should be empty by default")
	assert.Empty(t, cfg.Actions, "Actions should be empty by default")
}

// TestGetMinimalSandboxConfig verifies that GetMinimalSandboxConfig returns a valid config.
func TestGetMinimalSandboxConfig(t *testing.T) {
	cfg := GetMinimalSandboxConfig()

	require.NotNil(t, cfg, "GetMinimalSandboxConfig should return a config")
	assert.Equal(t, "latest", cfg.TerraformVersion, "TerraformVersion should be 'latest'")
	assert.NotNil(t, cfg.PublicRepo, "PublicRepo should be set")
	assert.Nil(t, cfg.ConnectedRepo, "ConnectedRepo should be nil for minimal config")
	assert.NotNil(t, cfg.EnvVarMap, "EnvVarMap should be initialized")
	assert.NotNil(t, cfg.VarsMap, "VarsMap should be initialized")

	// Verify public repo details
	assert.NotEmpty(t, cfg.PublicRepo.Repo, "PublicRepo.Repo should be set")
	assert.Equal(t, "/", cfg.PublicRepo.Directory, "PublicRepo.Directory should be '/'")
	assert.Equal(t, "main", cfg.PublicRepo.Branch, "PublicRepo.Branch should be 'main'")
}

// TestGetMinimalSandboxConfigWithConnectedRepo verifies the connected repo variant.
func TestGetMinimalSandboxConfigWithConnectedRepo(t *testing.T) {
	cfg := GetMinimalSandboxConfigWithConnectedRepo()

	require.NotNil(t, cfg, "GetMinimalSandboxConfigWithConnectedRepo should return a config")
	assert.Equal(t, "latest", cfg.TerraformVersion, "TerraformVersion should be 'latest'")
	assert.NotNil(t, cfg.ConnectedRepo, "ConnectedRepo should be set")
	assert.Nil(t, cfg.PublicRepo, "PublicRepo should be nil for connected repo config")

	// Verify connected repo details
	assert.NotEmpty(t, cfg.ConnectedRepo.Repo, "ConnectedRepo.Repo should be set")
	assert.Equal(t, "/", cfg.ConnectedRepo.Directory, "ConnectedRepo.Directory should be '/'")
	assert.Equal(t, "main", cfg.ConnectedRepo.Branch, "ConnectedRepo.Branch should be 'main'")
}

// TestGetMinimalRunnerConfig verifies that GetMinimalRunnerConfig returns a valid config.
func TestGetMinimalRunnerConfig(t *testing.T) {
	cfg := GetMinimalRunnerConfig()

	require.NotNil(t, cfg, "GetMinimalRunnerConfig should return a config")
	assert.Equal(t, "kubernetes", cfg.RunnerType, "RunnerType should be 'kubernetes'")
	assert.NotNil(t, cfg.EnvVarMap, "EnvVarMap should be initialized")
	assert.Empty(t, cfg.EnvVarMap, "EnvVarMap should be empty by default")
}
