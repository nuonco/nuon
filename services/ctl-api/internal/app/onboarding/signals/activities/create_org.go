package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	orgsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
func (a *Activities) createOnboardingOrg(ctx context.Context, accountID, orgName string) (*app.Org, error) {
	// Load account for CreateOrg (needs account for RBAC setup)
	var account app.Account
	if err := a.db.WithContext(ctx).First(&account, "id = ?", accountID).Error; err != nil {
		return nil, fmt.Errorf("unable to get account: %w", err)
	}

	// Set account context for BeforeCreate hooks (CreatedByID)
	ctx = cctx.SetAccountContext(ctx, &account)

	org, err := a.orgsHelpers.CreateOrg(ctx, &account, &orgshelpers.CreateOrgParams{
		Name:           orgName,
		UseSandboxMode: true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	// Set org context so event loop can start (startEventLoop calls signal.GetOrg which needs org in context)
	ctx = cctx.SetOrgContext(ctx, org)

	// Send v1 event loop signals (matching orgs/service/create_org.go)
	a.evClient.Send(ctx, org.ID, &orgsignals.Signal{
		Type: orgsignals.OperationCreated,
	})
	a.evClient.Send(ctx, org.ID, &orgsignals.Signal{
		Type: orgsignals.OperationProvision,
	})

	return org, nil
}
