package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

// fakeSandboxConfig is a faker provider that generates a minimal valid AppSandboxConfig.
// This provider is registered in init() and can be used via struct tags: `faker:"sandboxConfig"`
//
// The generated config uses a public repo configuration (simplest option for testing).
// TerraformVersion is required and set to "latest".
func fakeSandboxConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalSandboxConfig(), nil
}

// GetMinimalSandboxConfig returns a minimal valid AppSandboxConfig for use in tests.
//
// Uses PublicRepo configuration which doesn't require VCS connections.
// This is the simplest valid sandbox config for testing.
//
// Example usage:
//
//	sandbox := testseedconfig.GetMinimalSandboxConfig()
//	sandbox.DriftSchedule = generics.ToPtr("0 0 * * *")
func GetMinimalSandboxConfig() *config.AppSandboxConfig {
	return &config.AppSandboxConfig{
		TerraformVersion: "latest",
		PublicRepo: &config.PublicRepoConfig{
			Repo:      "https://github.com/nuonco/nuon-terraform-starter",
			Directory: "/",
			Branch:    "main",
		},
		EnvVarMap: map[string]string{},
		VarsMap:   map[string]string{},
	}
}

// GetMinimalSandboxConfigWithConnectedRepo returns a sandbox config using a connected repo.
// Use this when testing VCS connection functionality.
func GetMinimalSandboxConfigWithConnectedRepo() *config.AppSandboxConfig {
	return &config.AppSandboxConfig{
		TerraformVersion: "latest",
		ConnectedRepo: &config.ConnectedRepoConfig{
			Repo:      generics.GetFakeObj[string](),
			Directory: "/",
			Branch:    "main",
		},
		EnvVarMap: map[string]string{},
		VarsMap:   map[string]string{},
	}
}
