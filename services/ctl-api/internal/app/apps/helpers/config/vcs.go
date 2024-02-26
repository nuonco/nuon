package config

type PublicRepoConfig struct {
	Repo      string `mapstructure:"repo" toml:"repo"`
	Directory string `mapstructure:"directory" toml:"directory"`
	Branch    string `mapstructure:"branch" toml:"branch"`
}

type ConnectedRepoConfig struct {
	Repo      string `mapstructure:"repo" toml:"repo"`
	Directory string `mapstructure:"directory" toml:"repo"`
	Branch    string `mapstructure:"branch" toml:"branch"`
}
