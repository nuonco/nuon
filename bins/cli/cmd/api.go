package cmd

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
)

// newAPI creates an instance of the Nuon api client using the provided validator and config instance.
func newAPI(vld *validator.Validate, cfg *config.Config) (nuon.Client, error) {
	// Read config values from env and config file.
	// We intentionally do not read secrets from flags,
	// so we don't bind to pflags here even though we could.
	apiToken := cfg.GetString("api-token")
	orgID := cfg.GetString("org-id")
	apiURL := cfg.GetString("api-url")
	if apiURL == "" {
		apiURL = "https://ctl.prod.nuon.co"
	}

	api, err := nuon.New(vld,
		nuon.WithAuthToken(apiToken),
		nuon.WithOrgID(orgID),
		nuon.WithURL(apiURL),
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to init API client: %w", err)
	}
	return api, nil
}
