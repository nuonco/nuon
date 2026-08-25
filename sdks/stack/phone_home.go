package stack

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// PhoneHome sends install stack outputs to the control plane, marking the stack
// operation complete and updating install state.
//
// phoneHomeURL comes from Config.PhoneHomeURL rather than being composed here: it
// embeds a per-stack-version identifier, and routing it through the config response
// is what lets the Terraform module stop taking that identifier as an input.
func PhoneHome(ctx context.Context, opts Options, phoneHomeURL string, payload map[string]any) error {
	if strings.TrimSpace(phoneHomeURL) == "" {
		return fmt.Errorf("phone home: phone_home_url is required (read it from the stack config)")
	}
	if err := opts.validate(); err != nil {
		return fmt.Errorf("phone home: %w", err)
	}

	token, err := resolveToken(ctx, opts)
	if err != nil {
		return fmt.Errorf("phone home: %w", err)
	}

	client := newRunClient(runClientConfig{
		RunnerAPIURL: opts.APIURL,
		InstallID:    opts.InstallID,
		APIToken:     token,
		HTTPClient:   opts.HTTPClient,
	})

	if err := client.doWithRetry(ctx, http.MethodPost, phoneHomeURL, payload, nil); err != nil {
		return fmt.Errorf("phone home: %w", err)
	}

	return nil
}
