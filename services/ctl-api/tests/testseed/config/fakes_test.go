package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildMinimalAppConfig verifies that BuildMinimalAppConfig returns a valid config.
func TestBuildMinimalAppConfig(t *testing.T) {
	cfg := BuildMinimalAppConfig()

	require.NotNil(t, cfg, "BuildMinimalAppConfig should return a config")
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

// TestBuildMinimalSandboxConfig verifies that BuildMinimalSandboxConfig returns a valid config.
func TestBuildMinimalSandboxConfig(t *testing.T) {
	cfg := BuildMinimalSandboxConfig()

	require.NotNil(t, cfg, "BuildMinimalSandboxConfig should return a config")
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

// TestBuildMinimalSandboxConfigWithConnectedRepo verifies the connected repo variant.
func TestBuildMinimalSandboxConfigWithConnectedRepo(t *testing.T) {
	cfg := BuildMinimalSandboxConfigWithConnectedRepo()

	require.NotNil(t, cfg, "BuildMinimalSandboxConfigWithConnectedRepo should return a config")
	assert.Equal(t, "latest", cfg.TerraformVersion, "TerraformVersion should be 'latest'")
	assert.NotNil(t, cfg.ConnectedRepo, "ConnectedRepo should be set")
	assert.Nil(t, cfg.PublicRepo, "PublicRepo should be nil for connected repo config")

	// Verify connected repo details
	assert.NotEmpty(t, cfg.ConnectedRepo.Repo, "ConnectedRepo.Repo should be set")
	assert.Equal(t, "/", cfg.ConnectedRepo.Directory, "ConnectedRepo.Directory should be '/'")
	assert.Equal(t, "main", cfg.ConnectedRepo.Branch, "ConnectedRepo.Branch should be 'main'")
}

// TestBuildMinimalRunnerConfig verifies that BuildMinimalRunnerConfig returns a valid config.
func TestBuildMinimalRunnerConfig(t *testing.T) {
	cfg := BuildMinimalRunnerConfig()

	require.NotNil(t, cfg, "BuildMinimalRunnerConfig should return a config")
	assert.Equal(t, "kubernetes", cfg.RunnerType, "RunnerType should be 'kubernetes'")
	assert.NotNil(t, cfg.EnvVarMap, "EnvVarMap should be initialized")
	assert.Empty(t, cfg.EnvVarMap, "EnvVarMap should be empty by default")
}
