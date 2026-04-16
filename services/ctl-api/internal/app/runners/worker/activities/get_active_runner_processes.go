package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetActiveRunnerProcessesRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) GetActiveRunnerProcesses(ctx context.Context, req GetActiveRunnerProcessesRequest) ([]app.RunnerProcess, error) {
	var processes []app.RunnerProcess
	if res := a.db.WithContext(ctx).
		Where("runner_id = ? AND composite_status->>'status' IN ?", req.RunnerID, []string{
			string(app.RunnerProcessStatusActive),
			string(app.RunnerProcessStatusOffline),
		}).
		Find(&processes); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get active runner processes")
	}
	return processes, nil
}
