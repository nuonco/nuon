package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetStaleRunnerProcessesRequest struct {
	RunnerID    string `validate:"required"`
	ProcessType string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) GetStaleRunnerProcesses(ctx context.Context, req GetStaleRunnerProcessesRequest) ([]app.RunnerProcess, error) {
	var processes []app.RunnerProcess
	res := a.db.WithContext(ctx).
		Where("runner_id = ? AND type = ?", req.RunnerID, req.ProcessType).
		Order("created_at DESC").
		Offset(2).
		Find(&processes)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get stale runner processes: %w", res.Error)
	}

	return processes, nil
}
