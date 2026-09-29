package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
func (a *Activities) createOnboardingOrg(ctx context.Context, accountID, orgName string) (*app.Org, error) {
	var account app.Account
	if err := a.db.WithContext(ctx).First(&account, "id = ?", accountID).Error; err != nil {
		return nil, fmt.Errorf("unable to get account: %w", err)
	}

	ctx = cctx.SetAccountContext(ctx, &account)

	org, err := a.orgsHelpers.CreateOrg(ctx, &account, &orgshelpers.CreateOrgParams{
		Name:           orgName,
		UseSandboxMode: a.cfg.ForceOnboardingSandboxMode,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	return org, nil
}
