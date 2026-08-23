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

// EnsureInstallStackServiceAccount reconciles the service account an install stack
// authenticates as when it reads its own configuration from the runner API.
//
// It mints no token: the credential is handed to a person to paste into a shell or
// a CI secret, so it is created on demand from the dashboard. The OIDC path needs
// none at all.
//
// Keyed on the InstallStack, not the InstallStackVersion, which regenerates on every
// config change and would orphan the account and its tokens on every sync.
//
// Convergent, so a provision, a reprovision, and a retry all reach the same place —
// including stacks created before this account existed.
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

	// The stack reads install config and reports outputs back, which has no narrower
	// role than org admin today. Re-applied every call so older stacks pick it up.
	if err := a.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, stack.OrgID, acct.ID); err != nil {
		return nil, fmt.Errorf("unable to grant stack service account org role: %w", err)
	}

	return &EnsureInstallStackServiceAccountResponse{AccountID: acct.ID}, nil
}
