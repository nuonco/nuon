package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type HasOpenProcessShutdownRequest struct {
	RunnerID string `validate:"required"`
}

type HasOpenProcessShutdownResponse struct {
	HasOpenShutdown bool
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) HasOpenProcessShutdown(ctx context.Context, req HasOpenProcessShutdownRequest) (*HasOpenProcessShutdownResponse, error) {
	var count int64
	res := a.db.WithContext(ctx).
		Model(&app.RunnerProcessShutdown{}).
		Joins("JOIN runner_processes ON runner_processes.id = runner_process_shutdowns.runner_process_id AND runner_processes.runner_id = ?", req.RunnerID).
		Where("runner_process_shutdowns.composite_status->>'status' IN ?", []string{
			string(app.RunnerProcessShutdownStatusRequested),
			string(app.RunnerProcessShutdownStatusInProgress),
		}).
		Count(&count)
	if res.Error != nil {
		return nil, res.Error
	}

	return &HasOpenProcessShutdownResponse{HasOpenShutdown: count > 0}, nil
}
