package config

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type DockerBuildComponentConfig struct {
	CommonComponentFields

	Dockerfile string `mapstructure:"dockerfile" jsonschema:"required"`

	EnvVarMap map[string]string `mapstructure:"env_vars,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"  jsonschema:"oneof_required=public_repo"`

	// NOTE: the following parameters are not supported in the provider
	// Target	     string		   `mapstructure:"target" toml:"target"`
	// BuildArgs []string		`mapstructure:"build_args" toml:"build_args"`
}

func (t *DockerBuildComponentConfig) Validate() error {
	return nil
}

func (t *DockerBuildComponentConfig) Parse() error {
	return nil
}
