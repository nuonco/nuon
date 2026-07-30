package auth

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// ExchangeOIDCToken exchanges an OIDC ID token for a short-lived Nuon API
// token using the org's OIDC trust policies. It uses a tokenless API client
// since the exchange endpoint is unauthenticated.
func (a *Service) ExchangeOIDCToken(ctx context.Context, oidcToken, orgID string) (*models.ServiceExchangeOIDCTokenResponse, error) {
	if orgID == "" {
		orgID = a.cfg.OrgID
	}
	if orgID == "" {
		return nil, fmt.Errorf("no org configured: pass --org-id, run `nuon orgs select`, or set NUON_ORG_ID")
	}
	if oidcToken == "" {
		return nil, fmt.Errorf("no OIDC token provided")
	}

	if err := a.updateAPIClient(a.cfg.APIURL, a.cfg); err != nil {
		return nil, fmt.Errorf("couldn't create API client: %w", err)
	}

	resp, err := a.api.ExchangeOIDCToken(ctx, &models.ServiceExchangeOIDCTokenRequest{
		OrgID: &orgID,
		Token: &oidcToken,
	})
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	return resp, nil
}
