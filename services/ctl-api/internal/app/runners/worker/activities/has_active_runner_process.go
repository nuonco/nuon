package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type HasActiveRunnerProcessRequest struct {
	RunnerID string `validate:"required"`
}

type HasActiveRunnerProcessResponse struct {
	HasActive bool
}

// @temporal-gen-v2 activity
// @max-retries 1
// @by-field RunnerID
func (a *Activities) HasActiveRunnerProcess(ctx context.Context, req HasActiveRunnerProcessRequest) (*HasActiveRunnerProcessResponse, error) {
	var count int64
	res := a.db.WithContext(ctx).
		Model(&app.RunnerProcess{}).
		Where(app.RunnerProcess{RunnerID: req.RunnerID}).
		Scopes(generics.WhereJSONBStatusIn("composite_status",
			string(app.RunnerProcessStatusActive),
			string(app.RunnerProcessStatusPendingShutdown),
			string(app.RunnerProcessStatusShuttingDown),
		)).
		Count(&count)
	if res.Error != nil {
		return nil, res.Error
	}

	return &HasActiveRunnerProcessResponse{HasActive: count > 0}, nil
}
