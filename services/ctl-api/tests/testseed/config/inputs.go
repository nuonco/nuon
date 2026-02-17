package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

// fakeInputConfig is a faker provider that generates a minimal valid AppInputConfig.
func fakeInputConfig(v reflect.Value) (interface{}, error) {
	return BuildMinimalInputConfig(), nil
}

// BuildMinimalInputConfig returns a minimal valid AppInputConfig for use in tests.
// Returns an empty config (no inputs or groups) which is valid.
//
// Example usage:
//
//	inputs := testseedconfig.BuildMinimalInputConfig()
//	inputs.Groups = append(inputs.Groups, BuildInputGroup("database"))
//	inputs.Inputs = append(inputs.Inputs, BuildInput("db_url", "database"))
func BuildMinimalInputConfig() *config.AppInputConfig {
	return &config.AppInputConfig{
		Inputs: []config.AppInput{},
		Groups: []config.AppInputGroup{},
	}
}

// BuildInputGroup returns a fake input group with the given name.
func BuildInputGroup(name string) config.AppInputGroup {
	return config.AppInputGroup{
		Name:        name,
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
	}
}

// BuildInput returns a fake input for the given group.
//
// Parameters:
//   - name: the input name (used for templating)
//   - group: the group this input belongs to
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

// BuildCompleteInputConfig returns an input config with sample groups and inputs.
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
