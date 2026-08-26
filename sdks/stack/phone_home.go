package stack

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// PhoneHome sends install stack outputs to the control plane.
// This will update the install state, and mark the stack operation as complete.
func PhoneHome(ctx context.Context, runnerAPIURL, installID, phoneHomeID string, payload map[string]any) error {
	if strings.TrimSpace(installID) == "" {
		return fmt.Errorf("phone home: install_id is required")
	}
	if strings.TrimSpace(phoneHomeID) == "" {
		return fmt.Errorf("phone home: phone_home_id is required")
	}
	client := newRunClient(runClientConfig{
		RunnerAPIURL: runnerAPIURL,
		PhoneHomeID:  phoneHomeID,
	})
	if err := client.phoneHome(ctx, installID, payload); err != nil {
		return fmt.Errorf("phone home: %w", err)
	}
	return nil
}

func (c *runClient) phoneHome(ctx context.Context, installID string, payload map[string]any) error {
	url := fmt.Sprintf(
		"%s/v1/installs/%s/phone-home/%s",
		strings.TrimSuffix(c.cfg.RunnerAPIURL, "/"),
		installID,
		c.cfg.PhoneHomeID,
	)
	return c.doWithRetry(ctx, http.MethodPost, url, payload, nil)
}
