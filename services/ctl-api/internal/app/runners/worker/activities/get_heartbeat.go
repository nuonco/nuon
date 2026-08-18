package activities

import (
	"context"

	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetHeartBeatRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @max-retries 1
// @by-field ID
func (a *Activities) GetHeartBeat(ctx context.Context, req GetHeartBeatRequest) (*app.RunnerHeartBeat, error) {
	runner := app.RunnerHeartBeat{}
	res := a.chDB.WithContext(ctx).
		First(&runner, "id = ?", req.ID)
	if res.Error != nil {
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get runner")
	}

	return &runner, nil
}
