package stack

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// PhoneHome reports the result of an install-stack run to ctl-api over the
// public phone-home endpoint. It feeds the Terraform provider's
// stack_phone_home resource, which lets an install-stacks module report run
// status through the provider instead of constructing the HTTP request itself.
//
// apiURL is the base URL up to but excluding /v1 (e.g. https://api.nuon.co).
// payload is the JSON body the module assembled; the caller is responsible for
// setting request_type and phone_home_type. The per-install phone_home_id in
// the path is the secret, mirroring FetchConfig.
func PhoneHome(ctx context.Context, apiURL, installID, phoneHomeID string, payload map[string]any) error {
	if strings.TrimSpace(installID) == "" {
		return fmt.Errorf("phone home: install_id is required")
	}
	if strings.TrimSpace(phoneHomeID) == "" {
		return fmt.Errorf("phone home: phone_home_id is required")
	}
	client := newRunClient(runClientConfig{
		CtlAPIURL:   apiURL,
		PhoneHomeID: phoneHomeID,
	})
	if err := client.phoneHome(ctx, installID, payload); err != nil {
		return fmt.Errorf("phone home: %w", err)
	}
	return nil
}

// phoneHome POSTs the run report to
// /v1/installs/{install_id}/phone-home/{phone_home_id}.
func (c *runClient) phoneHome(ctx context.Context, installID string, payload map[string]any) error {
	url := fmt.Sprintf(
		"%s/v1/installs/%s/phone-home/%s",
		strings.TrimSuffix(c.cfg.CtlAPIURL, "/"),
		installID,
		c.cfg.PhoneHomeID,
	)
	return c.doWithRetry(ctx, http.MethodPost, url, payload, nil)
}
