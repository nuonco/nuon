package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateAccountRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateAccount(ctx context.Context, req CreateAccountRequest) (*app.Account, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org id from context")
	}

	acct, err := a.acctClient.FindAccount(ctx, account.ServiceAccountEmail(req.RunnerID))
	if err == nil {
		// NOTE(jm): each runner needs to be reprovisioned to properly create their roles, and then this should
		// be removed.
		a.authzClient.AddAccountOrgRole(ctx, app.RoleTypeRunner, orgID, acct.ID)
		return acct, nil
	}

	acct, err = a.acctClient.CreateServiceAccount(ctx, req.RunnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to create service account: %w", err)
	}

	a.authzClient.AddAccountOrgRole(ctx, app.RoleTypeRunner, orgID, acct.ID)
	return acct, nil
}
