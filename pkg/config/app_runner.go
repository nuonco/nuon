package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"

	"github.com/nuonco/nuon/pkg/config/source"
)

type AppRunnerConfig struct {
	Source string `mapstructure:"source,omitempty" toml:"source,omitempty"`

	RunnerType string `mapstructure:"runner_type" toml:"runner_type" jsonschema:"required"`

	EnvVarMap map[string]string `mapstructure:"env_vars,omitempty" toml:"env_vars,omitempty"`

	HelmDriver string `mapstructure:"helm_driver,omitempty" toml:"helm_driver,omitempty"`

	InitScriptURL string `mapstructure:"init_script_url" toml:"init_script_url"`

	// Deprecated
	EnvVars []EnvironmentVariable `mapstructure:"env_var,omitempty" toml:"env_var"`
}

func (a AppRunnerConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("source").Short("external configuration source").
		Long("Path to an external file containing runner configuration (YAML, JSON, or TOML)").
		Field("runner_type").Short("type of runner").Required().
		Long("Specifies how the runner executes deployments").
		Example("kubernetes").
		Example("docker").
		Example("vm").
		Field("env_vars").Short("environment variables").
		Long("Map of environment variables to pass to the runner as key-value pairs").
		Example("DEBUG=true").
		Example("LOG_LEVEL=info").
		Field("helm_driver").Short("Helm driver configuration").
		Long("Specifies the backend driver for Helm operations (e.g., 'configmap', 'secret')").
		Example("configmap").
		Example("secret").
		Field("init_script_url").Short("initialization script URL").
		Long("URL to a script that runs during runner initialization. Supports HTTP(S), git, file, and relative paths (./). Examples: https://example.com/script.sh, ./scripts/init.sh, git::https://github.com/org/repo//script.sh, file:///path/to/script.sh")
}

func (a *AppRunnerConfig) parse() error {
	if a == nil {
		return ErrConfig{
			Description: "an app runner config is required",
		}
	}

	if len(a.EnvVars) > 0 {
		return ErrConfig{
			Description: "env_var arrays are deprecated, please use env_vars map instead",
		}
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
