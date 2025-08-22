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
	SkipStatusSync    bool
}

// @temporal-gen activity
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

	// update InstallSandbox
	if !req.SkipStatusSync {
		run := app.InstallSandboxRun{}
		res = a.db.WithContext(ctx).
			Where("id = ?", req.RunID).
			First(&run)
		if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("unable to get install deploy: %w", res.Error)
		}

		if res.Error == gorm.ErrRecordNotFound {
			return nil
		}

		installSandbox := app.InstallSandbox{}
		resSandbox := a.db.WithContext(ctx).
			Where("install_id = ?", run.InstallID).
			First(&installSandbox)
		if resSandbox.Error != nil && resSandbox.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("unable to get install sandbox: %w", resSandbox.Error)
		}

		if resSandbox.Error == gorm.ErrRecordNotFound {
			return nil
		}

		res = a.db.WithContext(ctx).Model(&installSandbox).Updates(app.InstallSandbox{
			Status:            app.SandboxRunStatusToInstallSandboxStatus(req.Status),
			StatusDescription: req.StatusDescription,
		})
		if res.Error != nil {
			return fmt.Errorf("unable to update install sandbox: %w", res.Error)
		}
	}
	return nil
}
