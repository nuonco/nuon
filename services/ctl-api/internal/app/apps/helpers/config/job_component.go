package config

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type JobComponentConfig struct {
	Name         string                `mapstructure:"name"`
	Dependencies []string              `mapstructure:"dependencies"`
	ImageURL     string                `mapstructure:"image_url"`
	Tag          string                `mapstructure:"tag"`
	Cmd          []string              `mapstructure:"cmd"`
	EnvVars      []EnvironmentVariable `mapstructure:"env_var"`
	Args         []string              `mapstructure:"args"`
}

func (t *JobComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	return resource, nil
}
