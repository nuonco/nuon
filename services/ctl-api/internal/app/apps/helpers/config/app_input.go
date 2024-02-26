package config

type AppInput struct {
	Name        string `mapstructure:"name" toml:"name"`
	Description string `mapstructure:"description" toml:"description"`
	Default     string `mapstructure:"default" toml:"default"`
	Required    bool   `mapstructure:"required" toml:"required"`
	DisplayName string `mapstructure:"display_name" toml:"display_name"`
	Sensitive   bool   `mapstructure:"sensitive" toml:"sensitive"`
}

type AppInputConfig struct {
	Inputs []AppInput `mapstructure:"input" toml:"input"`
}

func (t *AppInputConfig) ToResourceType() string {
	return "nuon_app_input"
}

func (t *AppInputConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	return nestWithName("input", resource), nil
}
