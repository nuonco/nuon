package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// config holds config values, read from the `~/.nuon` config file and env vars.
type Config struct {
	*viper.Viper
}

// newConfig creates a new config instance.
func NewConfig() (*Config, error) {
	// Initialize Config instance
	cfg := &Config{viper.New()}

	// Read values from config file.
	cfg.SetConfigType("yaml")
	cfg.SetConfigName(".nuon")
	cfg.AddConfigPath("$HOME")
	if err := cfg.ReadInConfig(); err != nil {
		// The config file is optional, so we want to ignore "ConfigFileNotFoundError", but return all other errors.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Read values from env vars.
	cfg.SetEnvPrefix("NUON")
	cfg.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	cfg.AutomaticEnv()

	return cfg, nil
}

// BindCobraFlags binds config values to the flags of the provided cobra command.
func (cfg *Config) BindCobraFlags(cmd *cobra.Command) error {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		name := strings.ReplaceAll(f.Name, "-", "_")
		if !f.Changed && cfg.IsSet(name) {
			val := cfg.Get(name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
	return nil
}

// bindConfigFunc is an adapter enabling cobra commands to call config.BindFlags.
type BindCobraFunc func(*cobra.Command) error
