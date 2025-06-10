package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateRunStatusV2Request struct {
	RunID             string               `validate:"required"`
	Status            app.SandboxRunStatus `validate:"required"`
	StatusDescription string               `validate:"required"`
	SkipStatusSync    bool
}

// @temporal-gen activity
func (a *Activities) UpdateRunStatusV2(ctx context.Context, req UpdateRunStatusV2Request) error {
	install := app.InstallSandboxRun{
		ID: req.RunID,
	}

	compStatus := app.NewCompositeStatus(ctx, app.Status(req.Status))
	compStatus.StatusHumanDescription = req.StatusDescription
	res := a.db.WithContext(ctx).Model(&install).Updates(app.InstallSandboxRun{
		StatusV2: compStatus,
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

		if run.InstallSandboxID == nil {
			return nil
		}

		installSandbox := app.InstallSandbox{
			ID: *run.InstallSandboxID,
		}
		res = a.db.WithContext(ctx).Model(&installSandbox).Updates(app.InstallSandbox{
			StatusV2: compStatus,
		})
		if res.Error != nil {
			return fmt.Errorf("unable to update install sandbox: %w", res.Error)
		}
	}
	return nil
}
