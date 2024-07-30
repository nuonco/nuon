package config

type PublicRepoConfig struct {
	Repo      string `mapstructure:"repo,omitempty" jsonschema:"required"`
	Directory string `mapstructure:"directory,omitempty" jsonschema:"required"`
	Branch    string `mapstructure:"branch,omitempty" jsonschema:"required"`
}

type ConnectedRepoConfig struct {
	Repo      string `mapstructure:"repo,omitempty" jsonschema:"required"`
	Directory string `mapstructure:"directory,omitempty" jsonschema:"required"`
	Branch    string `mapstructure:"branch,omitempty" jsonschema:"required"`
}
