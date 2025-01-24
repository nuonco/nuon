package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateInstallWorkflowRunStatusRequest struct {
	RunID             string                             `validate:"required"`
	Status            app.InstallActionWorkflowRunStatus `validate:"required"`
	StatusDescription string                             `validate:"required"`
}

// @temporal-gen activity
// @by-id RunID
func (a *Activities) UpdateInstallWorkflowRunStatus(ctx context.Context, req UpdateInstallWorkflowRunStatusRequest) error {
	install := app.InstallActionWorkflowRun{
		ID: req.RunID,
	}
	res := a.db.WithContext(ctx).Model(&install).Updates(app.InstallActionWorkflowRun{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install action workflow run: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no run found: %s %w", req.RunID, gorm.ErrRecordNotFound)
	}
	return nil
}
