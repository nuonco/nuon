package cmd

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
)

const (
	API_TOKEN_ENV_NAME = "NUON_API_TOKEN"
	API_URL_ENV_NAME   = "NUON_API_URL"
	ORG_ID_ENV_NAME    = "NUON_ORG_ID"
)

func newAPI(vld *validator.Validate) (nuon.Client, error) {
	cfg, err := loadAPIConfig()
	if err != nil {
		return nil, fmt.Errorf("Missing API client config: %w", err)
	}
	api, err := nuon.New(vld,
		nuon.WithAuthToken(cfg.API_TOKEN),
		nuon.WithURL(cfg.API_URL),
		nuon.WithOrgID(cfg.ORG_ID),
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to init API client: %w", err)
	}
	return api, nil
}

type APIConfig struct {
	API_TOKEN string
	API_URL   string
	ORG_ID    string
}

// loadAPIConfig loads the config for the API from env vars, separately from the CLI flags. We don't want to support passing in secrets via flag.
// TODO(ja): Update this to use viper. Would be nice to support reading these from a config file as well as env vars.
func loadAPIConfig() (APIConfig, error) {
	config := APIConfig{}

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
