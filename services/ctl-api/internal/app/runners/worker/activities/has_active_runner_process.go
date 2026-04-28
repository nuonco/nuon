package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type HasActiveRunnerProcessRequest struct {
	RunnerID string `validate:"required"`
}

type HasActiveRunnerProcessResponse struct {
	HasActive bool
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) HasActiveRunnerProcess(ctx context.Context, req HasActiveRunnerProcessRequest) (*HasActiveRunnerProcessResponse, error) {
	var count int64
	res := a.db.WithContext(ctx).
		Model(&app.RunnerProcess{}).
		Where("runner_id = ? AND composite_status->>'status' = ?", req.RunnerID, string(app.RunnerProcessStatusActive)).
		Count(&count)
	if res.Error != nil {
		return nil, res.Error
	}

	return &HasActiveRunnerProcessResponse{HasActive: count > 0}, nil
}
