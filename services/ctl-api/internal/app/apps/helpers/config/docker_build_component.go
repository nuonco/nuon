package config

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type DockerBuildComponentConfig struct {
	Name         string   `mapstructure:"name"`
	Dependencies []string `mapstructure:"dependencies"`
	Dockerfile   string   `mapstructure:"dockerfile"`

	EnvVars []EnvironmentVariable `mapstructure:"env_var"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo"`

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

	return resource, nil
}
