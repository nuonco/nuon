package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func fakeInputConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalInputConfig(), nil
}

func GetMinimalInputConfig() *config.AppInputConfig {
	return &config.AppInputConfig{
		Inputs: []config.AppInput{},
		Groups: []config.AppInputGroup{},
	}
}

func GetInputGroup(name string) config.AppInputGroup {
	return config.AppInputGroup{
		Name:        name,
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
	}
}

func GetInput(name, group string) config.AppInput {
	return config.AppInput{
		Name:             name,
		DisplayName:      generics.GetFakeObj[string](),
		Description:      generics.GetFakeObj[string](),
		Group:            group,
		Type:             "string",
		Required:         false,
		Sensitive:        false,
		Internal:         false,
		UserConfigurable: true,
	}
}

func GetCompleteInputConfig() *config.AppInputConfig {
	return &config.AppInputConfig{
		Groups: []config.AppInputGroup{
			GetInputGroup("database"),
			GetInputGroup("api"),
		},
		Inputs: []config.AppInput{
			GetInput("db_host", "database"),
			GetInput("db_port", "database"),
			GetInput("api_key", "api"),
		},
	}
}
