package config

type AppSandboxConfig struct {
	TerraformVersion string               `mapstructure:"terraform_version" toml:"terraform_version"`
	ConnectedRepo    *ConnectedRepoConfig `mapstructure:"connected_repo" toml:"connected_repo"`
	PublicRepo       *PublicRepoConfig    `mapstructure:"public_repo" toml:"public_repo"`

	Vars []TerraformVariable `mapstructure:"var,omitempty" toml:"var"`
}

func (t *AppSandboxConfig) ToResourceType() string {
	return "nuon_app_sandbox"
}

func (t *AppSandboxConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	return nestWithName("sandbox", resource), nil
}
