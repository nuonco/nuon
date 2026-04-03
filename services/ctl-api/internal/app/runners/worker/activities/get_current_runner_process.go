package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetCurrentRunnerProcessRequest struct {
	RunnerID    string `validate:"required"`
	ProcessType string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) GetCurrentRunnerProcess(ctx context.Context, req GetCurrentRunnerProcessRequest) (*app.RunnerProcess, error) {
	var process app.RunnerProcess
	res := a.db.WithContext(ctx).
		Where("runner_id = ? AND type = ? AND composite_status->>'status' = ?", req.RunnerID, req.ProcessType, string(app.RunnerProcessStatusActive)).
		Preload("Shutdowns").
		Order("created_at DESC").
		First(&process)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get current runner process: %w", res.Error)
	}

	return &process, nil
}
