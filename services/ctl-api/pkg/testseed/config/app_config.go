package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

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
