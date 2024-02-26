package config

type TerraformVariable struct {
	Name  string `mapstructure:"name" toml:"name"`
	Value string `mapstructure:"value" toml:"value"`
}

type EnvironmentVariable struct {
	Name  string `mapstructure:"name" toml:"name"`
	Value string `mapstructure:"value" toml:"value"`
}
