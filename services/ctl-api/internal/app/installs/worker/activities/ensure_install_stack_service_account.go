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
// It deliberately does not mint a token. The credential is created on demand from
// the dashboard, by the same modal every other service account uses, because it is
// handed to a person to paste into a shell or a CI secret — and a token nobody has
// asked for yet is a credential sitting in the database with no owner watching it.
// The cost is that a stack has no usable token until someone creates one; that is
// the intended flow, and the OIDC path needs no token at all.
//
// The account is keyed on the InstallStack, not the InstallStackVersion. A stack
// regenerates its version on every config change, and re-keying per version would
// orphan the account — and any token issued against it — on every sync.
//
// Convergent, not a create hook: desired state is derived from Postgres on every
// call, so a provision, a reprovision, and a retry after a partial failure all reach
// the same place. That matters most for stacks created before this account existed,
// which pick one up on their next reconcile.
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

	// The stack reads install configuration and reports outputs back, which today has
	// no narrower role than org admin. Re-applied on every call so a stack created
	// before the role existed picks it up on its next reconcile.
	if err := a.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, stack.OrgID, acct.ID); err != nil {
		return nil, fmt.Errorf("unable to grant stack service account org role: %w", err)
	}

	return &EnsureInstallStackServiceAccountResponse{AccountID: acct.ID}, nil
}
