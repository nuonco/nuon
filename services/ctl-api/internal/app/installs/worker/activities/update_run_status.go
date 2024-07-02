package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateRunStatusRequest struct {
	RunID             string               `validate:"required"`
	Status            app.SandboxRunStatus `validate:"required"`
	StatusDescription string               `validate:"required"`
}

func (a *Activities) UpdateRunStatus(ctx context.Context, req UpdateRunStatusRequest) error {
	install := app.InstallSandboxRun{
		ID: req.RunID,
	}
	res := a.db.WithContext(ctx).Model(&install).Updates(app.InstallSandboxRun{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install run: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no run found: %s %w", req.RunID, gorm.ErrRecordNotFound)
	}
	return nil
}
