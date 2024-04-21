package config

type TerraformVariable struct {
	Name  string `mapstructure:"name,omitempty" toml:"name"`
	Value string `mapstructure:"value,omitempty" toml:"value"`
}

type EnvironmentVariable struct {
	Name  string `mapstructure:"name,omitempty" toml:"name"`
	Value string `mapstructure:"value,omitempty" toml:"value"`
}
