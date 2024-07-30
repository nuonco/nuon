package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

type AppRunnerConfig struct {
	Source string `mapstructure:"source,omitempty"`

	RunnerType string `mapstructure:"runner_type" jsonschema:"required"`

	EnvVarMap map[string]string `mapstructure:"env_vars,omitempty"`

	// Deprecated
	EnvVars []EnvironmentVariable `mapstructure:"env_var,omitempty" toml:"env_var"`
}

func (a *AppRunnerConfig) parse(ctx ConfigContext) error {
	if len(a.EnvVars) > 0 {
		return ErrConfig{
			Description: "env_var arrays are deprecated, please use env_vars map instead",
		}
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
