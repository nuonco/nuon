package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetMostRecentHeartBeatRequest struct {
	RunnerID string                `validate:"required"`
	Process  app.RunnerProcessType `json:"process,omitempty"`
}

// @temporal-gen-v2 activity
// @max-retries 5
// @by-field RunnerID
func (a *Activities) GetMostRecentHeartBeatRequest(ctx context.Context, req GetMostRecentHeartBeatRequest) (*app.RunnerHeartBeat, error) {
	hb, err := a.getMostRecentHeartBeat(ctx, req.RunnerID, req.Process)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner heart beat: %w", err)
	}

	return hb, nil
}

func (a *Activities) getMostRecentHeartBeat(ctx context.Context, runnerID string, process app.RunnerProcessType) (*app.RunnerHeartBeat, error) {
	if process != "" {
		hb, err := a.queryHeartBeat(ctx, runnerID, process)
		if err != nil {
			return nil, err
		}
		if hb != nil {
			return hb, nil
		}

		// TODO: remove this fallback once all runners send the correct process
		return a.queryHeartBeat(ctx, runnerID, app.RunnerProcessTypeUnknown)
	}

	return a.queryHeartBeat(ctx, runnerID, "")
}

func (a *Activities) queryHeartBeat(ctx context.Context, runnerID string, process app.RunnerProcessType) (*app.RunnerHeartBeat, error) {
	var latest []*app.LatestRunnerHeartBeat
	db := a.chDB.WithContext(ctx).
		Where("runner_id = ?", runnerID)
	if process != "" {
		db = db.Where("process = ?", process)
	}

	res := db.
		Order("created_at_latest desc").
		Limit(1).
		Find(&latest)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get heart beats")
	}
	if len(latest) == 0 {
		return nil, nil
	}

	return &app.RunnerHeartBeat{
		RunnerID:  latest[0].RunnerID,
		ProcessID: latest[0].ProcessID,
		Process:   latest[0].Process,
		Version:   latest[0].Version,
		AliveTime: latest[0].AliveTime,
		CreatedAt: latest[0].CreatedAt,
		StartedAt: latest[0].StartedAt,
	}, nil
}
