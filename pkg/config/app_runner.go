package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/config/source"
)

type AppRunnerConfig struct {
	Source string `mapstructure:"source,omitempty"`

	RunnerType string `mapstructure:"runner_type,omitempty" toml:"runner_type"`

	EnvironmentVariables []EnvironmentVariable `mapstructure:"env_var,omitempty" toml:"env_var"`
	EnvVarMap            map[string]string     `mapstructure:"-" toml:"env_vars"`
}

func (a *AppRunnerConfig) ToResourceType() string {
	return "nuon_app_runner"
}

func (a *AppRunnerConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(a)
	if err != nil {
		return nil, err
	}

	resource["app_id"] = "${var.app_id}"
	delete(resource, "source")

	return nestWithName("runner", resource), nil
}

func (a *AppRunnerConfig) parse(ctx ConfigContext) error {
	for k, v := range a.EnvVarMap {
		a.EnvironmentVariables = append(a.EnvironmentVariables, EnvironmentVariable{
			Name:  k,
			Value: v,
		})
	}

	if ctx == ConfigContextConfigOnly {
		return nil
	}
	if a.Source == "" {
		return nil
	}

	obj, err := source.LoadSource(a.Source)
	if err != nil {
		return ErrConfig{
			Description: fmt.Sprintf("unable to load source %s", a.Source),
			Err:         err,
		}
	}

	if err := mapstructure.Decode(obj, &a); err != nil {
		return err
	}
	return nil
}
