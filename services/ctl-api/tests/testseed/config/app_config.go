package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

// fakeAppConfig is a faker provider that generates a minimal valid AppConfig.
// This provider is registered in init() and can be used via struct tags: `faker:"appConfig"`
//
// The generated config includes:
// - Required fields: Version, Sandbox, Runner
// - Optional fields: DisplayName, Description
// - Empty slices for Components and Actions (can be added in tests)
func fakeAppConfig(v reflect.Value) (interface{}, error) {
	return &config.AppConfig{
		Version:     "1",
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
		Sandbox:     BuildMinimalSandboxConfig(),
		Runner:      BuildMinimalRunnerConfig(),
		Components:  config.ComponentList{},
		Actions:     []*config.ActionConfig{},
	}, nil
}

// BuildMinimalAppConfig returns a minimal valid AppConfig for use in tests.
// This is the recommended way to get a fake AppConfig in tests.
//
// Example usage:
//
//	cfg := testseedconfig.BuildMinimalAppConfig()
//	// Customize as needed
//	cfg.Components = append(cfg.Components, myComponent)
func BuildMinimalAppConfig() *config.AppConfig {
	return &config.AppConfig{
		Version:     "1",
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
		Sandbox:     BuildMinimalSandboxConfig(),
		Runner:      BuildMinimalRunnerConfig(),
		Components:  config.ComponentList{},
		Actions:     []*config.ActionConfig{},
	}
}

// BuildFullAppConfig returns an AppConfig with one of every component type.
// This is useful for tests that need to verify behavior across all component types.
//
// Includes: terraform module, helm chart, docker build, kubernetes manifest, and external image.
//
// Example usage:
//
//	cfg := testseedconfig.BuildFullAppConfig()
//	// All 5 component types are present
//	assert.Len(t, cfg.Components, 5)
func BuildFullAppConfig() *config.AppConfig {
	return &config.AppConfig{
		Version:     "1",
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
		Sandbox:     BuildMinimalSandboxConfig(),
		Runner:      BuildMinimalRunnerConfig(),
		Components: config.ComponentList{
			BuildTerraformComponent("terraform-component"),
			BuildHelmComponent("helm-component"),
			BuildDockerBuildComponent("docker-build-component"),
			BuildKubernetesManifestComponent("k8s-manifest-component"),
			BuildExternalImageComponent("external-image-component"),
		},
		Actions: []*config.ActionConfig{},
	}
}
