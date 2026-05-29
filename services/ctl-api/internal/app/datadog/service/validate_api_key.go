package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	ddclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
)

// validateAPIKey is the shared "is this key real?" probe used by both the
// create flow (gate before persistence) and the Test endpoint (post-create
// re-check). Returns a stderr.ErrInvalidRequest for the two user-visible
// failure modes (bad key, DD rejected) so the handler can surface them as
// 400. Network / 5xx errors bubble up unwrapped.
func (s *service) validateAPIKey(ctx context.Context, site, apiKey string) error {
	baseURL := ddclient.ResolveSiteURL(site)
	valid, err := s.ddClient.Validate(ctx, baseURL, apiKey)
	if err != nil {
		// Surface DD's structured 401/403 as an invalid-request so the
		// dashboard can render "wrong API key" rather than a generic
		// 500. Other transports (network, 5xx) stay as 500-class.
		var apiErr *ddclient.APIError
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 401 || apiErr.StatusCode == 403) {
			return stderr.NewInvalidRequest(fmt.Errorf("datadog rejected the api key: %s", apiErr.Body))
		}
		return fmt.Errorf("validate datadog api key: %w", err)
	}
	if !valid {
		return stderr.NewInvalidRequest(fmt.Errorf("datadog reports the api key is invalid"))
	}
	return nil
}
