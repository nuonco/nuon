package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/api/client"
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	API_TOKEN_ENV_NAME = "NUON_API_TOKEN"
	API_URL_ENV_NAME   = "NUON_API_URL"
	APP_ID_ENV_NAME    = "NUON_APP_ID"
	ORG_ID_ENV_NAME    = "NUON_ORG_ID"
)

type Config struct {
	API_TOKEN string
	API_URL   string
	ORG_ID    string
	APP_ID    string
}

type cli struct {
	api client.Client
	cfg Config
	v   *validator.Validate
}

// TODO: This function is a WIP. May need to refactor how we're managing each command to use this.
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

func loadConfig() (Config, error) {
	config := Config{}

	token := os.Getenv(API_TOKEN_ENV_NAME)
	if token == "" {
		return config, fmt.Errorf("Please set a nuon API token using: %s", API_TOKEN_ENV_NAME)
	}
	config.API_TOKEN = token

	url := os.Getenv(API_URL_ENV_NAME)
	if url != "" {
		config.API_URL = url
	} else {
		config.API_URL = "https://ctl.prod.nuon.co"
	}

	orgID := os.Getenv(ORG_ID_ENV_NAME)
	if orgID == "" {
		return config, fmt.Errorf("Please set a nuon org ID using: %s", ORG_ID_ENV_NAME)
	}
	config.ORG_ID = orgID

	appID := os.Getenv(APP_ID_ENV_NAME)
	config.APP_ID = appID

	return config, nil
}

func Execute() {
	rootCmd := &cobra.Command{
		Use:          "nuonctl",
		SilenceUsage: true,
		// TODO(ja): PersistentPreRunE is only inherited by immediate child commands,
		// so we still have to set this on each subcommand, so that it's children
		// will inherit it. There are a couple ways we can refactor things to avoid
		// this, but for now I'm just copy/pasting.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
	}

	v := validator.New()

	uiLog, err := ui.New(v)
	if err != nil {
		log.Fatalf("unable to initialize ui: %s", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Missing env variables: %s", err)
	}

	api, err := client.New(v,
		client.WithAuthToken(cfg.API_TOKEN),
		client.WithURL(cfg.API_URL),
		client.WithOrgID(cfg.ORG_ID),
	)
	if err != nil {
		log.Fatalf("Unable to init API client: %s", err)
	}

	ctx := context.Background()
	ctx = ui.WithContext(ctx, uiLog)
	c := &cli{
		api: api,
		cfg: cfg,
		v:   v,
	}

	if err := c.init(rootCmd.Flags()); err != nil {
		log.Fatalf("unable to initialize cli: %s", err)
	}

	namespaces := map[string]func(context.Context) cobra.Command{
		"apps":       c.registerApps,
		"components": c.registerComponents,
		"context":    c.registerContext,
		"installs":   c.registerInstalls,
		"version":    c.registerVersion,
	}
	for _, fn := range namespaces {
		cmd := fn(ctx)
		rootCmd.AddCommand(&cmd)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}
