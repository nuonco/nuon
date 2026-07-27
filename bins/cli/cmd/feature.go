package cmd

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const featureServiceAccountsAndTokens = "service-accounts-and-tokens"

// orgFeatureEnabled reports whether the current org has a feature flag turned
// on. It fails closed: any uncertainty — not logged in, no org selected,
// offline, or an API error — returns false.
func (c *cli) orgFeatureEnabled(feature string) bool {
	org := c.currentOrg()
	if org == nil {
		return false
	}
	return org.Features[feature]
}

// currentOrg returns the selected org, fetching it at most once. It is
// best-effort: on any failure (no config, no token, no org selected, API
// error) it returns nil and caches that result.
func (c *cli) currentOrg() *models.AppOrg {
	if c.orgInit {
		return c.org
	}
	c.orgInit = true

	if c.cfg == nil {
		if err := c.initConfig(); err != nil {
			return nil
		}
	}
	if c.cfg.APIToken == "" || c.cfg.OrgID == "" {
		return nil
	}
	if c.apiClient == nil {
		if err := c.initAPIClient(); err != nil {
			return nil
		}
	}

	org, err := c.apiClient.GetOrg(context.Background())
	if err != nil || org == nil {
		return nil
	}
	c.org = org
	return c.org
}
