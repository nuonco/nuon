package config

type DockerBuildComponentConfig struct {
	Name         string   `mapstructure:"name" toml:"name"`
	Dependencies []string `mapstructure:"dependencies" toml:"-"`
	Dockerfile   string   `mapstructure:"dockerfile" toml:"dockerfile"`

	EnvVars []EnvironmentVariable `mapstructure:"env_vars" toml:"env_vars"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo" toml:"public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo" toml:"connected_repo"`

	// NOTE: the following parameters are not supported in the provider
	//Target	     string		   `mapstructure:"target" toml:"target"`
	//BuildArgs []string		`mapstructure:"build_args" toml:"build_args"`
}

func (t *DockerBuildComponentConfig) ToResource() (map[string]interface{}, error) {
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
