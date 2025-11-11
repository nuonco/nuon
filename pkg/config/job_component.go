package config

import (
	"github.com/invopop/jsonschema"
)

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type JobComponentConfig struct {
	ImageURL string   `mapstructure:"image_url" jsonschema:"required"`
	Tag      string   `mapstructure:"tag" jsonschema:"required"`
	Cmd      []string `mapstructure:"cmd"`

	EnvVarMap map[string]string `mapstructure:"env_vars,omitempty"`
	Args      []string          `mapstructure:"args,omitempty"`

	// deprecated
	EnvVars []EnvironmentVariable `mapstructure:"env_var,omitempty"`
}

func (j JobComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("source").Short("source path or URL").
		Long("Optional source path or URL for the component configuration. Supports HTTP(S) URLs, git repositories, file paths, and relative paths (./). Examples: https://example.com/config.yaml, git::https://github.com/org/repo//config.yaml, file:///path/to/config.yaml, ./local/config.yaml").
		Field("type").Short("component type").
		Field("name").Short("component name").
		Field("var_name").Short("variable name for component output").
		Long("Optional name to use when storing component outputs as variables. If not specified, uses the component name").
		Field("dependencies").Short("component dependencies").
		Long("List of other components that must be deployed before this component. Automatically extracted from template references").
		Field("image_url").Short("job container image URL").Required().
		Long("Docker image URL to execute. Can be a public Docker registry or private registry").
		Example("ubuntu:22.04").
		Example("ghcr.io/myorg/myjob:v1.0.0").
		Field("tag").Short("image tag").Required().
		Long("Docker image tag to use").
		Example("latest").
		Example("v1.0.0").
		Field("cmd").Short("command to execute").
		Long("Command to run in the job container").
		Example("python").
		Example("bash").
		Field("env_vars").Short("environment variables").
		Long("Map of environment variables to pass to the job container").
		Field("args").Short("command arguments").
		Long("Arguments to pass to the command").
		Example("-c 'echo hello'").
		Example("script.py")
}

func (t *JobComponentConfig) Validate() error {
	if len(t.EnvVars) > 0 {
		return ErrConfig{
			Description: "the env_var array is deprecated, please use env_vars instead.",
		}
	}

	return nil
}

func (t *JobComponentConfig) Parse() error {
	return nil
}
