package config

type PublicRepoConfig struct {
	Repo      string `mapstructure:"repo,omitempty" toml:"repo"`
	Directory string `mapstructure:"directory,omitempty" toml:"directory"`
	Branch    string `mapstructure:"branch,omitempty" toml:"branch"`
}

type ConnectedRepoConfig struct {
	Repo      string `mapstructure:"repo,omitempty" toml:"repo"`
	Directory string `mapstructure:"directory,omitempty" toml:"repo"`
	Branch    string `mapstructure:"branch,omitempty" toml:"branch"`
}
