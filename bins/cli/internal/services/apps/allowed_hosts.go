package apps

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/urlpolicy"
	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// allowedHostsPolicy fetches where the control plane allows the app's config
// URLs to point. The check is advisory on the client: the server enforces the
// same policy at sync. So when the policy cannot be fetched, for example from
// a control plane that predates the endpoint, validation carries on without it.
func (s *Service) allowedHostsPolicy(ctx context.Context, appID string) *urlpolicy.Policy {
	resp, err := s.api.GetAppAllowedHosts(ctx, appID)
	if nuon.IsMissingRoute(err) {
		return nil
	}
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("skipped the allowed hosts check: unable to fetch the control plane's allowed hosts: %s", err))
		return nil
	}

	policy, err := allowedHostsPolicyFromResponse(resp)
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("skipped the allowed hosts check: %s", err))
		return nil
	}
	return policy
}

func allowedHostsPolicyFromResponse(resp *models.ServiceAppAllowedHosts) (*urlpolicy.Policy, error) {
	if resp == nil {
		return nil, nil
	}

	entries := make(map[string][]string, len(resp.Categories))
	for name, category := range resp.Categories {
		if category.Restricted {
			entries[name] = category.Entries
		}
	}
	return urlpolicy.FromEffective(urlpolicy.Mode(resp.Mode), entries)
}
