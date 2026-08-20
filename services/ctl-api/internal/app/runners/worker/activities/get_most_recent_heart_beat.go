package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// runner_heart_beats is ordered by (runner_id, process_id, created_at), so a created_at floor
// keeps this to a granule read instead of scanning every beat the runner has sent. Live runners
// beat every few seconds, so anything older than this means the runner is gone, not quiet.
const heartBeatLookback = time.Hour

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
	var beats []*app.RunnerHeartBeat
	db := a.chDB.WithContext(ctx).
		Where("runner_id = ?", runnerID).
		Where("created_at > ?", time.Now().Add(-heartBeatLookback))
	if process != "" {
		db = db.Where("process = ?", process)
	}

	res := db.
		Order("created_at desc").
		Limit(1).
		Find(&beats)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get heart beats")
	}
	if len(beats) == 0 {
		return nil, nil
	}

	hb := beats[0]
	hb.StartedAt = hb.CreatedAt.Add(-1 * hb.AliveTime)

	return hb, nil
}
