package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateSandboxBuildLogStreamRequest struct {
	AppSandboxBuildID string `validate:"required"`
	OrgID             string `validate:"required"`
	CreatedByID       string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateSandboxBuildLogStream(ctx context.Context, req CreateSandboxBuildLogStreamRequest) (*app.LogStream, error) {
	ctx = cctx.SetOrgIDContext(ctx, req.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, req.CreatedByID)

	ls := app.LogStream{
		OwnerType: "app_sandbox_builds",
		OwnerID:   req.AppSandboxBuildID,
		Open:      true,
	}

	if res := a.db.WithContext(ctx).Create(&ls); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create log stream")
	}

	svcAcct, err := a.acctClient.CreateServiceAccount(ctx, ls.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create service account")
	}

	a.authzClient.AddAccountOrgRole(ctx, app.RoleTypeRunner, req.OrgID, svcAcct.ID)

	token, err := a.acctClient.CreateToken(ctx, svcAcct.Email, defaultLogStreamTokenDuration)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create token")
	}

	ls.WriteToken = token.Token
	ls.RunnerAPIURL = a.cfg.RunnerAPIURL
	return &ls, nil
}
