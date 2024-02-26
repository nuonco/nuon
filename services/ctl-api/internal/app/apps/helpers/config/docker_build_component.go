package config

type DockerBuildComponentConfig struct {
	Dockerfile string                `mapstructure:"dockerfile" toml:"dockerfile"`
	Target     string                `mapstructure:"target" toml:"target"`
	BuildArgs  []string              `mapstructure:"build_args" toml:"build_args"`
	EnvVars    []EnvironmentVariable `mapstructure:"env_vars" toml:"env_vars"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo" toml:"public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo" toml:"connected_repo"`
}

func (t *DockerBuildComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	// blocks in terraform are singular, and repeated but in our config are plural.
	resource["env_var"] = resource["env_vars"]
	delete(resource, "env_var")

	return resource, nil
}
