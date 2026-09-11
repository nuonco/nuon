package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type GetLogStreamRequest struct {
	LogStreamID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field LogStreamID
func (a *Activities) GetLogStream(ctx context.Context, req GetLogStreamRequest) (*app.LogStream, error) {
	return a.getLogStream(ctx, req.LogStreamID)
}

func (a *Activities) getLogStream(ctx context.Context, logStreamID string) (*app.LogStream, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get org id from context: %w", err)
	}

	installLogStream := app.LogStream{}
	res := a.db.WithContext(ctx).Where("id = ? AND org_id = ?", logStreamID, orgID).First(&installLogStream)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	token, err := a.acctClient.CreateToken(ctx, account.ServiceAccountEmail(installLogStream.ID), defaultLogStreamTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("unable to create log stream token: %w", err)
	}

	installLogStream.RunnerAPIURL = a.cfg.RunnerAPIURL
	installLogStream.WriteToken = token.Token
	return &installLogStream, nil
}
