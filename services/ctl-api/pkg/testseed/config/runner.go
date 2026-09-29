package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
)

func fakeRunnerConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalRunnerConfig(), nil
}

func GetMinimalRunnerConfig() *config.AppRunnerConfig {
	return &config.AppRunnerConfig{
		RunnerType: "kubernetes",
		EnvVarMap:  map[string]string{},
	}
}
