package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/api/client"
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
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

func loadConfig() (Config, error) {
	config := Config{}

	token := os.Getenv(API_TOKEN_ENV_NAME)
	if token == "" {
		return config, fmt.Errorf("Please set a nuon API token using: %s", API_TOKEN_ENV_NAME)
	}
	config.API_TOKEN = token

	url := os.Getenv(API_URL_ENV_NAME)
	if url == "" {
		return config, fmt.Errorf("Please set a nuon API url using: %s", API_URL_ENV_NAME)
	}
	config.API_URL = url

	orgID := os.Getenv(ORG_ID_ENV_NAME)
	if orgID == "" {
		return config, fmt.Errorf("Please set a nuon org ID using: %s", ORG_ID_ENV_NAME)
	}
	config.ORG_ID = orgID

	appID := os.Getenv(APP_ID_ENV_NAME)
	if appID == "" {
		return config, fmt.Errorf("Please set a nuon app ID using: %s", APP_ID_ENV_NAME)
	}
	config.APP_ID = appID

	return config, nil
}

func Execute() {
	rootCmd := &cobra.Command{
		Use:          "nuonctl",
		SilenceUsage: true,
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
		"installs":  c.registerInstalls,
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
