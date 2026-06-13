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
	var processIDs []string
	res := a.db.WithContext(ctx).
		Model(&app.RunnerProcess{}).
		Where(app.RunnerProcess{RunnerID: req.RunnerID}).
		Where("composite_status->>'status' IN ?", []string{
			string(app.RunnerProcessStatusActive),
			string(app.RunnerProcessStatusPendingShutdown),
			string(app.RunnerProcessStatusShuttingDown),
		}).
		Pluck("id", &processIDs)
	if res.Error != nil {
		return nil, res.Error
	}

	if len(processIDs) == 0 {
		return &HasOpenProcessShutdownResponse{HasOpenShutdown: false}, nil
	}

	var count int64
	res = a.db.WithContext(ctx).
		Model(&app.RunnerProcessShutdown{}).
		Where("runner_process_id IN ?", processIDs).
		Where("composite_status->>'status' IN ?", []string{
			string(app.RunnerProcessShutdownStatusRequested),
			string(app.RunnerProcessShutdownStatusInProgress),
		}).
		Count(&count)
	if res.Error != nil {
		return nil, res.Error
	}

	return &HasOpenProcessShutdownResponse{HasOpenShutdown: count > 0}, nil
}
