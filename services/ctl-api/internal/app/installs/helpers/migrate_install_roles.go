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

	// build desired roles keyed by name
	desiredByName := make(map[string]app.AppAWSIAMRoleConfig, len(permCfg.Roles))
	for _, role := range permCfg.Roles {
		desiredByName[role.Name] = role
	}

	for _, install := range installs {
		// build existing roles keyed by their AppRoleConfig.Name
		existingByName := make(map[string]app.InstallRoles, len(install.InstallRoles))
		for _, ir := range install.InstallRoles {
			existingByName[ir.AppRoleConfig.Name] = ir
		}

		// roles that exist in new config
		for name, newRoleCfg := range desiredByName {
			if existing, found := existingByName[name]; found {
				// role exists — update AppRoleConfigID to point to new config, preserve other properties
				if existing.AppRoleConfigID != newRoleCfg.ID {
					if err := txn.WithContext(ctx).
						Model(&app.InstallRoles{}).
						Where("id = ?", existing.ID).
						Update("app_role_config_id", newRoleCfg.ID).Error; err != nil {
						return fmt.Errorf("unable to update install role %s for install %s: %w", existing.ID, install.ID, err)
					}
				}
			} else {
				// new role — create with default properties
				newRole := app.InstallRoles{
					InstallID:       install.ID,
					AppRoleConfigID: newRoleCfg.ID,
				}
				if err := txn.WithContext(ctx).Create(&newRole).Error; err != nil {
					return fmt.Errorf("unable to create install role for install %s: %w", install.ID, err)
				} m
			}
		}

		// roles that exist in install but not in new config — soft delete
		for name, existing := range existingByName {
			if _, desired := desiredByName[name]; desired {
				continue
			}
			if err := txn.WithContext(ctx).Delete(&app.InstallRoles{}, "id = ?", existing.ID).Error; err != nil {
				return fmt.Errorf("unable to delete install role %s: %w", existing.ID, err)
			}
		}
	}

	return nil
}
