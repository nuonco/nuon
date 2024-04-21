package config

type AppSandboxConfig struct {
	TerraformVersion string               `mapstructure:"terraform_version,omitempty" toml:"terraform_version"`
	ConnectedRepo    *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo"`
	PublicRepo       *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo"`

	Vars   []TerraformVariable `mapstructure:"var,omitempty" toml:"var"`
	VarMap map[string]string   `mapstructure:"-" toml:"vars"`
}

func (t *AppSandboxConfig) ToResourceType() string {
	return "nuon_app_sandbox"
}

func (t *AppSandboxConfig) parse() error {
	for k, v := range t.VarMap {
		t.Vars = append(t.Vars, TerraformVariable{
			Name:  k,
			Value: v,
		})
	}
	return nil
}

func (t *AppSandboxConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	return nestWithName("sandbox", resource), nil
}
