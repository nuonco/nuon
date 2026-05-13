package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetLatestRunnerHeartBeatRequest struct {
	RunnerID    string                `validate:"required"`
	ProcessType app.RunnerProcessType `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) GetLatestRunnerHeartBeat(ctx context.Context, req GetLatestRunnerHeartBeatRequest) (*app.LatestRunnerHeartBeat, error) {
	var hb app.LatestRunnerHeartBeat
	res := a.chDB.WithContext(ctx).
		Where("runner_id = ? AND process = ?", req.RunnerID, req.ProcessType).
		Order("created_at_latest desc").
		First(&hb)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no heartbeat found for runner %s process %s", req.RunnerID, req.ProcessType)
		}

		return nil, fmt.Errorf("unable to get latest runner heartbeat: %w", res.Error)
	}

	return &hb, nil
}
