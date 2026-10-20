package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateAccountRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen-v2 activity
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
		if err := a.ensureRunnerInstallRole(ctx, orgID, req.RunnerID, acct.ID); err != nil {
			return nil, err
		}
		return acct, nil
	}

	acct, err = a.acctClient.CreateServiceAccount(ctx, req.RunnerID, "")
	if err != nil {
		return nil, fmt.Errorf("unable to create service account: %w", err)
	}

	a.authzClient.AddAccountOrgRole(ctx, app.RoleTypeRunner, orgID, acct.ID)
	if err := a.ensureRunnerInstallRole(ctx, orgID, req.RunnerID, acct.ID); err != nil {
		return nil, err
	}
	return acct, nil
}

func (a *Activities) ensureRunnerInstallRole(ctx context.Context, orgID, runnerID, accountID string) error {
	var runner app.Runner
	if err := a.db.WithContext(ctx).Preload("RunnerGroup").Where(app.Runner{ID: runnerID}).Take(&runner).Error; err != nil {
		return nil
	}
	if runner.RunnerGroup.Type != app.RunnerGroupTypeInstall || runner.RunnerGroup.OwnerID == "" {
		return nil
	}
	if err := a.authzClient.EnsureRunnerInstallRole(ctx, orgID, runner.RunnerGroup.OwnerID, accountID); err != nil {
		return fmt.Errorf("unable to ensure runner install role: %w", err)
	}
	return nil
}
