package config

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type JobComponentConfig struct {
	ImageURL string   `mapstructure:"image_url" jsonschema:"required"`
	Tag      string   `mapstructure:"tag" jsonschema:"required"`
	Cmd      []string `mapstructure:"cmd" jsonschema:"required"`

	EnvVarMap map[string]string `mapstructure:"env_vars,omitempty"`
	Args      []string          `mapstructure:"args,omitempty"`

	// deprecated
	EnvVars []EnvironmentVariable `mapstructure:"env_var,omitempty"`
}

func (t *JobComponentConfig) Validate() error {
	if len(t.EnvVars) > 0 {
		return ErrConfig{
			Description: "the env_var array is deprecated, please use env_vars instead.",
		}
	}

	return nil
}
