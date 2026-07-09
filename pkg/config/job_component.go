package config

import (
	"github.com/invopop/jsonschema"
)

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type JobComponentConfig struct {
	ImageURL string   `mapstructure:"image_url" toml:"image_url" jsonschema:"required"`
	Tag      string   `mapstructure:"tag" toml:"tag" jsonschema:"required"`
	Cmd      []string `mapstructure:"cmd" toml:"cmd"`

	EnvVarMap map[string]string `mapstructure:"env_vars,omitempty" toml:"env_vars,omitempty"`
	Args      []string          `mapstructure:"args,omitempty" toml:"args,omitempty"`

	BuildTimeout  string `mapstructure:"build_timeout,omitempty" toml:"build_timeout,omitempty" features:"template" nuonhash:"omitempty"`
	DeployTimeout string `mapstructure:"deploy_timeout,omitempty" toml:"deploy_timeout,omitempty" features:"template" nuonhash:"omitempty"`

	// deprecated
	EnvVars []EnvironmentVariable `mapstructure:"env_var,omitempty" toml:"env_var,omitempty"`
}

func (j JobComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
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
		Example("script.py").
		Field("build_timeout").Short("build operation timeout").
		Long("Duration string for build operations (e.g., \"30m\", \"1h\"). Default: 5m. Max: 1h").
		Default("5m").
		Example("30m").
		Example("1h").
		Field("deploy_timeout").Short("deploy operation timeout").
		Long("Duration string for job execution (e.g., \"30m\", \"1h\"). Default: 15m. Max: 1h").
		Default("15m").
		Example("30m").
		Example("1h")
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
