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
		Sandbox:     GetMinimalSandboxConfig(),
		Runner:      GetMinimalRunnerConfig(),
		Components:  config.ComponentList{},
		Actions:     []*config.ActionConfig{},
	}, nil
}

// GetMinimalAppConfig returns a minimal valid AppConfig for use in tests.
// This is the recommended way to get a fake AppConfig in tests.
//
// Example usage:
//
//	cfg := testseedconfig.GetMinimalAppConfig()
//	// Customize as needed
//	cfg.Components = append(cfg.Components, myComponent)
func GetMinimalAppConfig() *config.AppConfig {
	return &config.AppConfig{
		Version:     "1",
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
		Sandbox:     GetMinimalSandboxConfig(),
		Runner:      GetMinimalRunnerConfig(),
		Components:  config.ComponentList{},
		Actions:     []*config.ActionConfig{},
	}
}
