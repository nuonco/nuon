package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

const queueLogStreamTokenDuration = time.Hour

type HydrateLogStreamRequest struct {
	LogStreamID string `validate:"required"`
	OrgID       string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) HydrateLogStream(ctx context.Context, req *HydrateLogStreamRequest) (*app.LogStream, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, err
	}

	var stream app.LogStream
	res := a.db.WithContext(ctx).
		Where(app.LogStream{ID: req.LogStreamID, OrgID: req.OrgID}).
		First(&stream)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, fmt.Sprintf("unable to get log stream %s", req.LogStreamID))
	}

	token, err := a.acctClient.CreateToken(ctx, account.ServiceAccountEmail(stream.ID), queueLogStreamTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("create log stream token: %w", err)
	}

	stream.RunnerAPIURL = a.cfg.RunnerAPIURL
	stream.WriteToken = token.Token
	return &stream, nil
}
