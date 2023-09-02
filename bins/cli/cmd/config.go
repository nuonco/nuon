package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// bindConfig uses viper to read config values from env vars and config files, based on the flags defined in the cobra command.
func bindConfig(cmd *cobra.Command) error {
	// Read values from config file.
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(".nuon")
	v.AddConfigPath("$HOME")
	if err := v.ReadInConfig(); err != nil {
		// The config file is optional, so we want to ignore "ConfigFileNotFoundError", but return all other errors.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Read values from env vars.
	v.SetEnvPrefix("NUON")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// Bind cobra flags to viper config values.
	// If a flag is not set, this checks to see if a config value is set.
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		name := strings.ReplaceAll(f.Name, "-", "_")
		if !f.Changed && v.IsSet(name) {
			val := v.Get(name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})

	return nil
}
