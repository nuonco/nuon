package config

type AppRunnerConfig struct {
	RunnerType string `mapstructure:"runner_type,omitempty" toml:"runner_type"`

	EnvironmentVariables []EnvironmentVariable `mapstructure:"env_var,omitempty" toml:"env_var"`
	EnvVarMap            map[string]string     `mapstructure:"-" toml:"env_vars"`
}

func (a *AppRunnerConfig) ToResourceType() string {
	return "nuon_app_runner"
}

func (a *AppRunnerConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(a)
	if err != nil {
		return nil, err
	}

	resource["app_id"] = "${var.app_id}"

	return nestWithName("runner", resource), nil
}

func (a *AppRunnerConfig) parse() error {
	for k, v := range a.EnvVarMap {
		a.EnvironmentVariables = append(a.EnvironmentVariables, EnvironmentVariable{
			Name:  k,
			Value: v,
		})
	}
	return nil
}
