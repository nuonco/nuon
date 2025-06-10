package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateInstallWorkflowRunStatusV2Request struct {
	RunID             string                             `validate:"required"`
	Status            app.InstallActionWorkflowRunStatus `validate:"required"`
	StatusDescription string                             `validate:"required"`
}

// @temporal-gen activity
// @by-id RunID
func (a *Activities) UpdateInstallWorkflowRunStatusV2(ctx context.Context, req UpdateInstallWorkflowRunStatusV2Request) error {
	install := app.InstallActionWorkflowRun{
		ID: req.RunID,
	}

	compStatus := app.NewCompositeStatus(ctx, app.Status(req.Status))
	compStatus.StatusHumanDescription = req.StatusDescription
	res := a.db.WithContext(ctx).Model(&install).Updates(app.InstallActionWorkflowRun{
		StatusV2: compStatus,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install action workflow run: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no run found: %s %w", req.RunID, gorm.ErrRecordNotFound)
	}
	return nil
}
