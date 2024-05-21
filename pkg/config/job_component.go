package config

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type JobComponentConfig struct {
	Name         string                `mapstructure:"name,omitempty"`
	Dependencies []string              `mapstructure:"dependencies,omitempty"`
	ImageURL     string                `mapstructure:"image_url,omitempty"`
	Tag          string                `mapstructure:"tag,omitempty"`
	Cmd          []string              `mapstructure:"cmd,omitempty"`
	EnvVars      []EnvironmentVariable `mapstructure:"env_var,omitempty"`
	EnvVarMap    map[string]string     `mapstructure:"env_vars,omitempty"`
	Args         []string              `mapstructure:"args,omitempty"`
}

func (t *JobComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	delete(resource, "env_vars")
	return resource, nil
}

func (t *JobComponentConfig) parse(ConfigContext) error {
	for k, v := range t.EnvVarMap {
		t.EnvVars = append(t.EnvVars, EnvironmentVariable{
			Name:  k,
			Value: v,
		})
	}

	return nil
}
