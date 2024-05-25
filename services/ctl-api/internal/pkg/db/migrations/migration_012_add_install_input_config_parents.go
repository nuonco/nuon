package migrations

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (a *Migrations) migration012AddInstallInputConfigParents(ctx context.Context) error {
	var allInputs []*app.InstallInputs
	res := a.db.WithContext(ctx).
		Preload("Install").
		Preload("Install.App").
		Preload("Install.App.AppInputConfigs").
		Preload("Install.App.AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC")
		}).
		Find(&allInputs)
	if res.Error != nil {
		return res.Error
	}

	for _, inputs := range allInputs {
		// if an input is stored with no actual values, we remove it.
		if len(inputs.Values) < 1 || len(inputs.Install.App.AppInputConfigs) < 1 {
			a.l.Info("deleting install input config")
			res := a.db.WithContext(ctx).Delete(&app.InstallInputs{
				ID: inputs.ID,
			})
			if res.Error != nil {
				return fmt.Errorf("unable to delete input config with no app inputs: %w", res.Error)
			}
			continue
		}

		// update the pointer to point to the app input config parent
		a.l.Info("adding parent app input config to install input config")
		res := a.db.WithContext(ctx).
			Model(&app.InstallInputs{
				ID: inputs.ID,
			}).
			Updates(app.InstallInputs{AppInputConfigID: inputs.Install.App.AppInputConfigs[0].ID})
		if res.Error != nil {
			return fmt.Errorf("unable to update install input to point to app input config: %w", res.Error)
		}
	}

	// hard delete any install inputs that were deleted
	var deletedInstallInputs []app.InstallInputs
	res = a.db.WithContext(ctx).Unscoped().Find(&deletedInstallInputs)
	if res.Error != nil {
		return fmt.Errorf("unable to find deleted install inputs: %w", res.Error)
	}

	if len(deletedInstallInputs) < 1 {
		return nil
	}

	res = a.db.WithContext(ctx).Unscoped().Delete(&deletedInstallInputs)
	if res.Error != nil {
		return fmt.Errorf("unable to delete install inputs that were hard deleted: %w", res.Error)
	}

	return nil
}
