package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultFilePath         string        = "~/.nuon"
	defaultAPIURL           string        = "https://api.nuon.co"
	defaultConfigFileEnvVar string        = "NUON_CONFIG_FILE"
	defaultGitHubAppName    string        = "nuon-connect"
	defaultCleanupTimeout   time.Duration = time.Second * 2
)

// config holds config values, read from the `~/.nuon` config file and env vars.
type Config struct {
	*viper.Viper

	APIToken         string `mapstructure:"api_token"`
	APIURL           string `mapstructure:"api_url"`
	OrgID            string `mapstructure:"org_id"`
	DisableTelemetry bool   `mapstructure:"disable_telemetry"`
	Debug            bool   `mapstructure:"debug"`
	Preview          bool   `mapstructure:"preview"`

	// internal configuration, not designed to be used by users
	GitHubAppName   string        `mapstructure:"github_app_name"`
	Env             string        `mapstructure:"-"`
	CleanupTimeout  time.Duration `mapstructure:"-"`
	SegmentWriteKey string        `mapstructure:"-"`
	SentryDSN       string        `mapstructure:"-"`
	UserID          string        `mapstructure:"-"`
}

// newConfig creates a new config instance.
func NewConfig(customFilepath string) (*Config, error) {
	cfg := &Config{
		Viper:          viper.New(),
		APIURL:         defaultAPIURL,
		GitHubAppName:  defaultGitHubAppName,
		Debug:          Debug(),
		Preview:        Preview(),
		CleanupTimeout: defaultCleanupTimeout,
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
	if cfg.GetString("api_token") != "" {
		cfg.APIToken = cfg.GetString("api_token")
	}
	if cfg.GetString("api_url") != "" {
		cfg.APIURL = cfg.GetString("api_url")
	}
	if cfg.GetString("org_id") != "" {
		cfg.OrgID = cfg.GetString("org_id")
	}
	if cfg.GetString("github_app_name") != "" {
		cfg.GitHubAppName = cfg.GetString("github_app_name")
	}
	if cfg.GetBool("disable_telemetry") {
		cfg.DisableTelemetry = cfg.GetBool("disable_telemetry")
	}

	cfg.Env = cfg.envFromAPIURL(cfg.APIURL)
	cfg.SegmentWriteKey = cfg.segmentWriteKey(cfg.Env)
	cfg.SentryDSN = cfg.sentryDSN(cfg.Env)

	return cfg, nil
}

// readConfigFile reads config values from a yaml file at ~/.nuon
func (c *Config) readConfigFile(customFP string) error {
	cfgFP := defaultFilePath
	isDefault := true
	if customFP != "" && customFP != cfgFP {
		cfgFP = customFP
		isDefault = false
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
		if isDefault {
			return nil
		}

		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		if isDefault {
			return nil
		}

		return err
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
