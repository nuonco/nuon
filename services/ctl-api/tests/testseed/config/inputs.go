package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func fakeInputConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalInputConfig(), nil
}

func BuildMinimalInputConfig() *config.AppInputConfig {
	return &config.AppInputConfig{
		Inputs: []config.AppInput{},
		Groups: []config.AppInputGroup{},
	}
}

func BuildInputGroup(name string) config.AppInputGroup {
	return config.AppInputGroup{
		Name:        name,
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
	}
}

func BuildInput(name, group string) config.AppInput {
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

func BuildCompleteInputConfig() *config.AppInputConfig {
	return &config.AppInputConfig{
		Groups: []config.AppInputGroup{
			BuildInputGroup("database"),
			BuildInputGroup("api"),
		},
		Inputs: []config.AppInput{
			BuildInput("db_host", "database"),
			BuildInput("db_port", "database"),
			BuildInput("api_key", "api"),
		},
	}
}
