package config

type JobComponentConfig struct {
	Name         string                `mapstructure:"name" toml:"name"`
	Dependencies []string              `mapstructure:"dependencies" toml:"-"`
	ImageURL     string                `mapstructure:"image_url" toml:"image_url"`
	Tag          string                `mapstructure:"tag" toml:"tag"`
	Cmd          []string              `mapstructure:"cmd" toml:"cmd"`
	EnvVars      []EnvironmentVariable `mapstructure:"env_vars" toml:"env_vars"`
	Args         []string              `mapstructure:"args" toml:"args"`
}

func (t *JobComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	// blocks in terraform are singular, and repeated but in our config are plural.
	resource["env_var"] = resource["env_vars"]
	delete(resource, "env_vars")

	return resource, nil
}
