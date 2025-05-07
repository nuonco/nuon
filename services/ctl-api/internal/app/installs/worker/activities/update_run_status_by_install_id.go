package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateRunStatusByInstallIDRequest struct {
	InstallID         string               `validate:"required"`
	Status            app.SandboxRunStatus `validate:"required"`
	StatusDescription string               `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UpdateRunStatusByInstallID(ctx context.Context, req UpdateRunStatusByInstallIDRequest) error {
	latestRun := app.InstallSandboxRun{}
	res := a.db.WithContext(ctx).
		Where("install_id = ?", req.InstallID).
		Order("created_at desc").
		First(&latestRun)
	if res.Error != nil {
		return fmt.Errorf("unable to get install run: %w", res.Error)
	}

	install := app.InstallSandboxRun{
		ID: latestRun.ID,
	}
	res = a.db.WithContext(ctx).Model(&install).Updates(app.InstallSandboxRun{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install run: %w", res.Error)
	}

	if latestRun.InstallSandboxID == nil {
		return nil
	}

	installSandbox := app.InstallSandbox{
		ID: generics.FromPtrStr(latestRun.InstallSandboxID),
	}
	res = a.db.WithContext(ctx).Model(&installSandbox).Updates(app.InstallSandbox{
		Status:            app.SandboxRunStatusToInstallSandboxStatus(req.Status),
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install sandbox: %w", res.Error)
	}

	return nil
}
