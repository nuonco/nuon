package stack

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// PhoneHome reports install stack outputs, marking the operation complete. Takes
// phoneHomeURL from Config rather than composing it.
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
