package activities

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type UpdateInstallWorkflowRunStatusRequest struct {
	RunID             string                             `validate:"required"`
	Status            app.InstallActionWorkflowRunStatus `validate:"required"`
	StatusDescription string                             `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunID
func (a *Activities) UpdateInstallWorkflowRunStatus(ctx context.Context, req UpdateInstallWorkflowRunStatusRequest) error {
	install := app.InstallActionWorkflowRun{
		ID: req.RunID,
	}
	if req.Status == app.InstallActionRunStatusCancelled {
		_, err := statusactivities.CancelIfUnfinished(ctx, a.db, &install, req.StatusDescription, true)
		return err
	}
	query := a.db.WithContext(ctx).Model(&install).Where("status <> ?", app.InstallActionRunStatusCancelled)
	res := query.Updates(app.InstallActionWorkflowRun{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install action workflow run: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		var existing app.InstallActionWorkflowRun
		if err := a.db.WithContext(ctx).Where(app.InstallActionWorkflowRun{ID: req.RunID}).First(&existing).Error; err == nil && existing.Status == app.InstallActionRunStatusCancelled {
			return nil
		}
		return fmt.Errorf("no run found: %s %w", req.RunID, gorm.ErrRecordNotFound)
	}
	if req.Status == app.InstallActionRunStatusFinished {
		if err := a.helpers.RecordInstallActionWorkflowRunApplied(ctx, req.RunID, time.Now().UTC()); err != nil {
			return fmt.Errorf("unable to record applied action workflow run: %w", err)
		}
	}
	return nil
}
