package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Helpers) MigrateInstallRoles(ctx context.Context, txn *gorm.DB, appID string, permCfg app.AppPermissionsConfig) error {
	if permCfg.ID == "" {
		return nil
	}

	var installs []app.Install
	res := txn.WithContext(ctx).
		Preload("InstallRoles").
		Preload("InstallRoles.AppRoleConfig").
		Where("app_id = ?", appID).
		Find(&installs)
	if res.Error != nil {
		return fmt.Errorf("unable to get installs for app %s: %w", appID, res.Error)
	}

	for _, install := range installs {
		oldByName := make(map[string]app.InstallRoles, len(install.InstallRoles))
		for _, ir := range install.InstallRoles {
			oldByName[ir.AppRoleConfig.Name] = ir
		}

		// soft-delete all existing install roles
		for _, ir := range install.InstallRoles {
			if err := txn.WithContext(ctx).Delete(&app.InstallRoles{}, "id = ?", ir.ID).Error; err != nil {
				return fmt.Errorf("unable to delete install role %s: %w", ir.ID, err)
			}
		}

		// create fresh set from new config, carrying over properties from matching old roles
		for _, role := range permCfg.Roles {
			newRole := app.InstallRoles{
				InstallID:       install.ID,
				AppRoleConfigID: role.ID,
			}

			if old, found := oldByName[role.Name]; found {
				newRole.Enabled = old.Enabled
				newRole.Provisioned = old.Provisioned
				newRole.RoleID = old.RoleID
			}

			if err := txn.WithContext(ctx).Create(&newRole).Error; err != nil {
				return fmt.Errorf("unable to create install role for install %s: %w", install.ID, err)
			}
		}
	}

	return nil
}
