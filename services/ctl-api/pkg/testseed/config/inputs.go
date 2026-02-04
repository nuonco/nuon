package config

import (
	"reflect"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

// fakeInputConfig is a faker provider that generates a minimal valid AppInputConfig.
func fakeInputConfig(v reflect.Value) (interface{}, error) {
	return GetMinimalInputConfig(), nil
}

// GetMinimalInputConfig returns a minimal valid AppInputConfig for use in tests.
// Returns an empty config (no inputs or groups) which is valid.
//
// Example usage:
//
//	inputs := testseedconfig.GetMinimalInputConfig()
//	inputs.Groups = append(inputs.Groups, GetInputGroup("database"))
//	inputs.Inputs = append(inputs.Inputs, GetInput("db_url", "database"))
func GetMinimalInputConfig() *config.AppInputConfig {
	return &config.AppInputConfig{
		Inputs: []config.AppInput{},
		Groups: []config.AppInputGroup{},
	}
}

// GetInputGroup returns a fake input group with the given name.
func GetInputGroup(name string) config.AppInputGroup {
	return config.AppInputGroup{
		Name:        name,
		DisplayName: generics.GetFakeObj[string](),
		Description: generics.GetFakeObj[string](),
	}
}

// GetInput returns a fake input for the given group.
//
// Parameters:
//   - name: the input name (used for templating)
//   - group: the group this input belongs to
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

// GetCompleteInputConfig returns an input config with sample groups and inputs.
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
