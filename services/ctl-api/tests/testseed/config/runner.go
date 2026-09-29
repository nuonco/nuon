package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
)

func fakeRunnerConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalRunnerConfig(), nil
}

func BuildMinimalRunnerConfig() *config.AppRunnerConfig {
	return &config.AppRunnerConfig{
		RunnerType: "kubernetes",
		EnvVarMap:  map[string]string{},
	}
}
