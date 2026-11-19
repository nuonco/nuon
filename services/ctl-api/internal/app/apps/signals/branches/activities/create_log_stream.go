package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const defaultLogStreamTokenDuration time.Duration = time.Hour

type CreateLogStreamRequest struct {
	AppBranchRunID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateLogStream(ctx context.Context, req CreateLogStreamRequest) (*app.LogStream, error) {
	ls := app.LogStream{
		OwnerType: "app_branch_runs",
		OwnerID:   req.AppBranchRunID,
		Open:      true,
	}

	if res := a.db.WithContext(ctx).Create(&ls); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create log stream")
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org id from context")
	}

	token, err := a.acctClient.IssueLogStreamToken(ctx, orgID, ls.ID, defaultLogStreamTokenDuration)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create token")
	}

	ls.WriteToken = token.Token
	ls.RunnerAPIURL = a.cfg.RunnerAPIURL
	return &ls, nil
}
