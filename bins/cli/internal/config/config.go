package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultFilePath         string = "~/.nuon"
	defaultAPIURL           string = "https://ctl.prod.nuon.co"
	defaultConfigFileEnvVar string = "NUON_CONFIG_FILE"
)

// config holds config values, read from the `~/.nuon` config file and env vars.
type Config struct {
	*viper.Viper

	APIToken string `mapstructure:"api_token"`
	APIURL   string `mapstructure:"api_url"`
	OrgID    string `mapstructure:"org_id"`
}

// newConfig creates a new config instance.
func NewConfig(customFilepath string) (*Config, error) {
	cfg := &Config{
		Viper:  viper.New(),
		APIURL: defaultAPIURL,
	}

	// Read values from config file.
	if err := cfg.readConfigFile(customFilepath); err != nil {
		return nil, err
	}

	// Read values from env vars.
	cfg.SetEnvPrefix("NUON")
	cfg.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	cfg.AutomaticEnv()

	// Set global config values
	cfg.APIToken = cfg.GetString("api_token")
	cfg.APIURL = cfg.GetString("api_url")
	cfg.OrgID = cfg.GetString("org_id")

	return cfg, nil
}

// readConfigFile reads config values from a yaml file at ~/.nuon
func (c *Config) readConfigFile(customFP string) error {
	cfgFP := defaultFilePath
	if customFP != "" {
		cfgFP = customFP
	}
	if os.Getenv(defaultConfigFileEnvVar) != "" {
		cfgFP = os.Getenv(defaultConfigFileEnvVar)
	}

	var err error
	cfgFP, err = homedir.Expand(cfgFP)
	if err != nil {
		return fmt.Errorf("unable to expand home directory: %w", err)
	}

	c.SetConfigFile(cfgFP)
	c.SetConfigType("yaml")

	err = c.ReadInConfig()
	if err == nil {
		return nil
	}

	nfe := &viper.ConfigFileNotFoundError{}
	if errors.As(err, &nfe) {
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return fmt.Errorf("unable to load config file: %w", err)
}

// BindCobraFlags binds config values to the flags of the provided cobra command.
func (c *Config) BindCobraFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		name := strings.ReplaceAll(f.Name, "-", "_")
		if !f.Changed && c.IsSet(name) {
			val := c.Get(name)

			//nolint:all
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
