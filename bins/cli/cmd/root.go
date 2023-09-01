package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/bins/cli/internal/apps"
	"github.com/powertoolsdev/mono/bins/cli/internal/components"
	"github.com/powertoolsdev/mono/bins/cli/internal/installs"
	"github.com/powertoolsdev/mono/bins/cli/internal/orgs"
	"github.com/powertoolsdev/mono/bins/cli/internal/version"
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

func loadAPIConfig() (Config, error) {
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

	return config, nil
}

// Execute is essentially the init method of the CLI. It initializes all the components and composes them together.
func Execute() {
	// Create root command that all other commands will be nested under.
	// if there are any flags or other settings that we want to be "global",
	// they should be configured on the root command
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

	// Create an api client instance for the services to use.
	v := validator.New()
	cfg, err := loadAPIConfig()
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

	// Construct an instance of each domain service, injecting the API client,
	// and passing the service to the CLI commands that will use it.
	// TODO(ja): It looks like we can have a generic Service type these can all share,
	// and then we may be able to loop over each domain with the same code,
	// instead of copy/pasting.
	installsCmd := registerInstalls(installs.New(api))
	rootCmd.AddCommand(&installsCmd)

	appsCmd := registerApps(apps.New(api))
	rootCmd.AddCommand(&appsCmd)

	orgsCmd := registerOrgs(orgs.New(api))
	rootCmd.AddCommand(&orgsCmd)

	versionCmd := registerVersion(version.New(api))
	rootCmd.AddCommand(&versionCmd)

	componentsCmd := registerComponents(components.New(api))
	rootCmd.AddCommand(&componentsCmd)

	// Execute the root command with a context object.
	ctx := context.Background()

	// TODO(ja): Should this be in cmd? Will circle back.
	uiLog, err := ui.New(v)
	if err != nil {
		log.Fatalf("unable to initialize ui: %s", err)
	}
	ctx = ui.WithContext(ctx, uiLog)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(2)
	}
}
