package config

type TerraformVariablesFile struct {
	Contents string `toml:"contents" mapstructure:"contents,omitempty" features:"get,template"`
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type TerraformModuleComponentConfig struct {
	TerraformVersion string `mapstructure:"terraform_version" jsonschema:"required"`

	EnvVarMap      map[string]string        `mapstructure:"env_vars,omitempty"`
	VarsMap        map[string]string        `mapstructure:"vars,omitempty"`
	VariablesFiles []TerraformVariablesFile `mapstructure:"var_file,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"  jsonschema:"oneof_required=public_repo"`

	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`

	// deprecated
	Variables []TerraformVariable   `mapstructure:"var,omitempty" `
	EnvVars   []EnvironmentVariable `mapstructure:"env_var,omitempty"`
}

func (t *TerraformModuleComponentConfig) Parse() error {
	return nil
}

func (t *TerraformModuleComponentConfig) Validate() error {
	if len(t.Variables) > 0 {
		return ErrConfig{
			Description: "the var array is deprecated, please use vars instead.",
		}
	}
	if len(t.EnvVars) > 0 {
		return ErrConfig{
			Description: "the env_var array is deprecated, please use env_vars instead.",
		}
	}

	return nil
}
