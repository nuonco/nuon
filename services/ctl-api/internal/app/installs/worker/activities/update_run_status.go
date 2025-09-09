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

	run := app.InstallSandboxRun{}
	res = a.db.WithContext(ctx).
		Where("id = ?", req.RunID).
		First(&run)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("unable to get install sandbox run: %w", res.Error)
	}

	install2 := &app.Install{}
	res = a.db.WithContext(ctx).
		Where("id = ?", run.InstallID).
		First(install2)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("unable to get install: %w", res.Error)
	}

	installSandbox := app.InstallSandbox{}
	res = a.db.WithContext(ctx).
		Where("install_id = ?", install2.ID).
		First(&installSandbox)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("unable to get install sandbox: %w", res.Error)
	}

	if res.Error == gorm.ErrRecordNotFound {
		return nil
	}

	runUpdate := app.InstallSandboxRun{
		ID: req.RunID,
	}
	res = a.db.WithContext(ctx).Model(&runUpdate).Updates(app.InstallSandboxRun{
		InstallSandboxID: &installSandbox.ID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install sandbox run with sandbox ID: %w", res.Error)
	}

	// Only update status if SkipStatusSync is false
	if !req.SkipStatusSync {
		installSandbox = app.InstallSandbox{
			ID: installSandbox.ID,
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
