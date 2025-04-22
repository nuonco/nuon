package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateInstallSandboxStatusRequest struct {
	RunID             string                   `validate:"required"`
	Status            app.InstallSandboxStatus `validate:"required"`
	StatusDescription string                   `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UpdateInstallSandboxStatus(ctx context.Context, req UpdateInstallSandboxStatusRequest) error {
	run := app.InstallSandboxRun{}
	res := a.db.WithContext(ctx).
		Where("id = ?", req.RunID).
		First(&run)
	if res.Error != nil {
		return fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	if run.InstallSandboxID != nil {
		return nil
	}

	installSandbox := app.InstallSandbox{
		ID: *run.InstallSandboxID,
	}
	res = a.db.WithContext(ctx).Model(&installSandbox).Updates(app.InstallSandbox{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install sandbox: %w", res.Error)
	}

	// TODO: enable check after install sanbox backfill
	// if res.RowsAffected < 1 {
	// 	return fmt.Errorf("no install sandbox found: %s %w", req.InstallSandboxID, gorm.ErrRecordNotFound)
	// }

	return nil
}
