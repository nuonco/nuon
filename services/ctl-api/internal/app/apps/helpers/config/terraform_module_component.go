package config

type TerraformModuleComponentConfig struct {
	TerraformVersion string                `mapstructure:"terraform_version" toml:"terraform_version"`
	Variables        []TerraformVariable   `mapstructure:"vars" toml:"variables"`
	EnvVars          []EnvironmentVariable `mapstructure:"env_vars" toml:"env_vars"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo" toml:"public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo" toml:"connected_repo"`
}

func (t *TerraformModuleComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	// blocks in terraform are singular, and repeated but in our config are plural.
	resource["env_var"] = resource["env_vars"]
	delete(resource, "env_vars")

	resource["var"] = resource["vars"]
	delete(resource, "vars")

	return resource, nil
}
