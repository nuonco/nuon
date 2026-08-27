package stack

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Options configures a call against the runner API's stacks namespace. Only APIURL
// and InstallID are required.
type Options struct {
	// APIURL is the runner API base URL, e.g. https://runner.nuon.co.
	APIURL string

	// InstallID identifies the install. Not a secret; authentication authorizes the read.
	InstallID string

	// APIToken authenticates the caller. Optional; see Options for the fallbacks.
	APIToken string

	// OrgID is required only on the OIDC path, where the exchange has to name the
	// org whose trust policies apply. Falls back to NUON_ORG_ID.
	OrgID string

	HTTPClient *http.Client
}

func (o Options) validate() error {
	if strings.TrimSpace(o.APIURL) == "" {
		return fmt.Errorf("api_url is required")
	}
	if strings.TrimSpace(o.InstallID) == "" {
		return fmt.Errorf("install_id is required")
	}

	return nil
}

// FetchConfig reads an install's rendered stack config. The returned Config carries
// PhoneHomeURL, so a later report needs no other identifier.
func FetchConfig(ctx context.Context, opts Options) (*Config, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	token, err := resolveToken(ctx, opts)
	if err != nil {
		return nil, err
	}

	client := newRunClient(runClientConfig{
		RunnerAPIURL: opts.APIURL,
		InstallID:    opts.InstallID,
		APIToken:     token,
		HTTPClient:   opts.HTTPClient,
	})

	cfg, err := client.fetchConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch stack config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("fetch stack config: runner api returned no config block")
	}

	return cfg, nil
}
