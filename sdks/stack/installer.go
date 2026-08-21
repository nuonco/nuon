package stack

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Options configures a call against the runner API's stacks namespace.
//
// APIURL and InstallID are the only required fields. Credentials are resolved in the
// same order the Nuon CLI uses — an explicit APIToken, then NUON_API_TOKEN, then an
// ambient OIDC token exchanged for a short-lived one — so a customer applying from
// CI never has to hold a long-lived secret.
type Options struct {
	// APIURL is the runner API base URL, e.g. https://runner.nuon.co.
	APIURL string

	// InstallID identifies the install whose stack config is being read. Not a
	// secret: authentication is what authorizes the read.
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

// FetchConfig reads the rendered install-stack config for an install.
//
// The returned Config carries PhoneHomeURL, so a caller that later reports outputs
// does not need to know the phone-home ID: it arrives over this authenticated
// channel instead of being rendered into the customer's Terraform variables.
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
