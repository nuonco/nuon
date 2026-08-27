package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type EnsureInstallStackServiceAccountRequest struct {
	InstallStackID string `json:"install_stack_id" validate:"required"`
}

type EnsureInstallStackServiceAccountResponse struct {
	AccountID string `json:"account_id"`
}

// EnsureInstallStackServiceAccount reconciles the account an install stack
// authenticates as. Mints no token — those are created on demand from the dashboard.
//
// Keyed on the InstallStack, not the version, which regenerates on every config
// change. Convergent, so it also backfills older stacks.
//
// @temporal-gen-v2 activity
// @by-field InstallStackID
// @start-to-close-timeout 2m
func (a *Activities) EnsureInstallStackServiceAccount(
	ctx context.Context, req *EnsureInstallStackServiceAccountRequest,
) (*EnsureInstallStackServiceAccountResponse, error) {
	if err := a.v.StructCtx(ctx, req); err != nil {
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	var stack app.InstallStack
	if res := a.db.WithContext(ctx).
		Where(app.InstallStack{ID: req.InstallStackID}).
		First(&stack); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to load install stack: %w")
	}

	acct, err := a.acctClient.EnsureServiceAccount(ctx, stack.ID, "")
	if err != nil {
		return nil, fmt.Errorf("unable to ensure stack service account: %w", err)
	}

	// Grant before revoke: a crash in between leaves the account over-privileged
	// rather than locked out, and the next call converges.
	if err := a.authzClient.EnsureStackInstallRole(ctx, stack.OrgID, stack.InstallID, acct.ID); err != nil {
		return nil, fmt.Errorf("unable to ensure stack service account role: %w", err)
	}

	// Stacks provisioned before the scoped role existed hold org admin.
	if err := a.authzClient.RemoveAccountOrgRoleByType(ctx, app.RoleTypeOrgAdmin, stack.OrgID, acct.ID); err != nil {
		return nil, fmt.Errorf("unable to revoke stack service account org admin: %w", err)
	}

	return &EnsureInstallStackServiceAccountResponse{AccountID: acct.ID}, nil
}
