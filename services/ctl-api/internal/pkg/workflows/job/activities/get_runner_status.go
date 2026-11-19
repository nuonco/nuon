package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetRunnerStatusRequest struct {
	ID string `validate:"required"`
}

// PkgWorkflowsJobGetRunnerStatus returns the current status of the given runner.
// A missing runner returns app.RunnerStatusUnknown rather than an error, since
// callers only use this to decide dispatch strategy and a missing runner just
// means it has not been provisioned yet.
//
// @temporal-gen-v2 activity
// @by-field ID
// @start-to-close-timeout 1m
func (a *Activities) PkgWorkflowsJobGetRunnerStatus(ctx context.Context, req *GetRunnerStatusRequest) (app.RunnerStatus, error) {
	runner := app.Runner{}
	if res := a.db.WithContext(ctx).
		Where(&app.Runner{ID: req.ID}).
		First(&runner); res.Error != nil {
		return app.RunnerStatusUnknown, nil
	}

	return runner.Status, nil
}
