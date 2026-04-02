package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetMostRecentHeartBeatRequest struct {
	RunnerID string            `validate:"required"`
	Process  app.RunnerProcess `json:"process,omitempty"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) GetMostRecentHeartBeatRequest(ctx context.Context, req GetMostRecentHeartBeatRequest) (*app.RunnerHeartBeat, error) {
	hb, err := a.getMostRecentHeartBeat(ctx, req.RunnerID, req.Process)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner heart beat: %w", err)
	}

	return hb, nil
}

func (a *Activities) getMostRecentHeartBeat(ctx context.Context, runnerID string, process app.RunnerProcess) (*app.RunnerHeartBeat, error) {
	if process != "" {
		hb, err := a.queryHeartBeat(ctx, runnerID, process)
		if err != nil {
			return nil, err
		}
		if hb != nil {
			return hb, nil
		}

		// TODO: remove this fallback once all runners send the correct process
		return a.queryHeartBeat(ctx, runnerID, app.RunnerProcessUknown)
	}

	return a.queryHeartBeat(ctx, runnerID, "")
}

func (a *Activities) queryHeartBeat(ctx context.Context, runnerID string, process app.RunnerProcess) (*app.RunnerHeartBeat, error) {
	var hb app.RunnerHeartBeat
	db := a.chDB.WithContext(ctx).
		Where("runner_id = ?", runnerID)
	if process != "" {
		db = db.Where("process = ?", process)
	}
	res := db.
		Order("created_at desc").
		Limit(1).
		First(&hb)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrap(res.Error, "unable to get heart beats")
	}

	return &hb, nil
}
