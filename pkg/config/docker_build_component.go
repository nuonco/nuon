package config

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type DockerBuildComponentConfig struct {
	Name         string   `mapstructure:"name,omitempty"`
	Dependencies []string `mapstructure:"dependencies,omitempty"`
	Dockerfile   string   `mapstructure:"dockerfile,omitempty"`

	EnvVars   []EnvironmentVariable `mapstructure:"env_var,omitempty"`
	EnvVarMap map[string]string     `mapstructure:"env_vars,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"`

	// NOTE: the following parameters are not supported in the provider
	//Target	     string		   `mapstructure:"target" toml:"target"`
	//BuildArgs []string		`mapstructure:"build_args" toml:"build_args"`
}

func (t *DockerBuildComponentConfig) parse(ConfigContext) error {
	for k, v := range t.EnvVarMap {
		t.EnvVars = append(t.EnvVars, EnvironmentVariable{
			Name:  k,
			Value: v,
		})
	}

	return nil
}

func (t *DockerBuildComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	delete(resource, "env_vars")
	return resource, nil
}
