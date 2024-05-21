package config

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type TerraformModuleComponentConfig struct {
	Name             string                `mapstructure:"name,omitempty"`
	TerraformVersion string                `mapstructure:"terraform_version,omitempty"`
	Dependencies     []string              `mapstructure:"dependencies,omitempty"`
	Variables        []TerraformVariable   `mapstructure:"var,omitempty"`
	VarsMap          map[string]string     `mapstructure:"vars,omitempty"`
	EnvVars          []EnvironmentVariable `mapstructure:"env_var,omitempty"`
	EnvVarMap        map[string]string     `mapstructure:"env_vars,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo"`
}

func (t *TerraformModuleComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"
	delete(resource, "env_vars")
	delete(resource, "vars")

	return resource, nil
}

func (t *TerraformModuleComponentConfig) parse(ctx ConfigContext) error {
	for k, v := range t.EnvVarMap {
		t.EnvVars = append(t.EnvVars, EnvironmentVariable{
			Name:  k,
			Value: v,
		})
	}

	for k, v := range t.VarsMap {
		t.Variables = append(t.Variables, TerraformVariable{
			Name:  k,
			Value: v,
		})
	}

	return nil
}
