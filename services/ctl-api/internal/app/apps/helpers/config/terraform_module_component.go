package config

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type TerraformModuleComponentConfig struct {
	Name		 string		       `mapstructure:"name"`
	TerraformVersion string		       `mapstructure:"terraform_version"`
	Dependencies	 []string	       `mapstructure:"dependencies"`
	Variables	 []TerraformVariable   `mapstructure:"var"`
	EnvVars		 []EnvironmentVariable `mapstructure:"env_var"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo" toml:"public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo" toml:"connected_repo"`
}

func (t *TerraformModuleComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	return resource, nil
}
