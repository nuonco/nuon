package config

type AppInput struct {
	Name        string `mapstructure:"name,omitempty" toml:"name"`
	Description string `mapstructure:"description,omitempty" toml:"description"`
	Default     string `mapstructure:"default" toml:"default"`
	Required    bool   `mapstructure:"required" toml:"required"`
	DisplayName string `mapstructure:"display_name,omitempty" toml:"display_name"`
	Sensitive   bool   `mapstructure:"sensitive" toml:"sensitive"`
}

type AppInputConfig struct {
	Inputs []AppInput `mapstructure:"input,omitempty" toml:"input"`
}

func (a *AppInputConfig) ToResourceType() string {
	return "nuon_app_input"
}

func (a *AppInputConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(a)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	return nestWithName("input", resource), nil
}

func (a *AppInputConfig) parse() error {
	return nil
}
