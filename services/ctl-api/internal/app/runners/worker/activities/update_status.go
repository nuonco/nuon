package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/processjobupdates"
)

type UpdateStatusRequest struct {
	RunnerID          string           `validate:"required"`
	Status            app.RunnerStatus `validate:"required"`
	StatusDescription string           `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	runner := app.Runner{
		ID: req.RunnerID,
	}
	res := a.db.WithContext(ctx).Model(&runner).Updates(app.Runner{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update runner status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no runner found: %s %w", req.RunnerID, gorm.ErrRecordNotFound)
	}

	// Fan out a runner_status_changed update to any in-flight ProcessJob
	// workflows for this runner so they can react without polling.
	a.helpers.UpdateProcessJobsForRunner(ctx, req.RunnerID, processjobupdates.UpdateNameRunnerStatusChanged, processjobupdates.RunnerStatusChangedPayload{
		RunnerID: req.RunnerID,
		Status:   string(req.Status),
	})

	return nil
}
