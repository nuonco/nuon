package config

type AppRunnerConfig struct {
	RunnerType string `mapstructure:"runner_type" toml:"runner_type"`

	EnvironmentVariables []EnvironmentVariable `mapstructure:"env_var" toml:"env_var"`
}

func (t *AppRunnerConfig) ToResourceType() string {
	return "nuon_app_runner"
}

func (t *AppRunnerConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}

	resource["app_id"] = "${var.app_id}"

	return nestWithName("runner", resource), nil
}
