package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
)

// fakeRunnerConfig is a faker provider that generates a minimal valid AppRunnerConfig.
// This provider is registered in init() and can be used via struct tags: `faker:"runnerConfig"`
//
// The generated config uses the "kubernetes" runner type (most common).
func fakeRunnerConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalRunnerConfig(), nil
}

// BuildMinimalRunnerConfig returns a minimal valid AppRunnerConfig for use in tests.
//
// Uses "kubernetes" runner type which is the most common in production.
//
// Example usage:
//
//	runner := testseedconfig.BuildMinimalRunnerConfig()
//	runner.EnvVarMap["DEBUG"] = "true"
func BuildMinimalRunnerConfig() *config.AppRunnerConfig {
	return &config.AppRunnerConfig{
		RunnerType: "kubernetes",
		EnvVarMap:  map[string]string{},
	}
}
