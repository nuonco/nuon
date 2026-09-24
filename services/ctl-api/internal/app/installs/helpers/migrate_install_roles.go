package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Install roles mirror the identities that exist in the customer account, so each
// install keeps one live set. Rows are repointed in place: a stable ID is what keeps
// install_role_usage history and enabled/provisioned/role_id attached across syncs.
func (s *Helpers) MigrateInstallRoles(ctx context.Context, txn *gorm.DB, appID string, permCfg app.AppPermissionsConfig) error {
	if permCfg.ID == "" {
		return nil
	}

	var installs []app.Install
	res := txn.WithContext(ctx).
		Preload("InstallRoles").
		Preload("InstallRoles.AppRoleConfig").
		Where(app.Install{AppID: appID}).
		Find(&installs)
	if res.Error != nil {
		return fmt.Errorf("unable to get installs for app %s: %w", appID, res.Error)
	}

	for _, install := range installs {
		staleByName := make(map[string]app.InstallRoles, len(install.InstallRoles))
		for _, ir := range install.InstallRoles {
			staleByName[ir.AppRoleConfig.Name] = ir
		}

		for _, role := range permCfg.Roles {
			existing, found := staleByName[role.Name]
			if !found {
				newRole := app.InstallRoles{
					InstallID:       install.ID,
					AppRoleConfigID: role.ID,
				}
				if err := txn.WithContext(ctx).Create(&newRole).Error; err != nil {
					return fmt.Errorf("unable to create install role for install %s: %w", install.ID, err)
				}
				continue
			}
			delete(staleByName, role.Name)

			if existing.AppRoleConfigID == role.ID {
				continue
			}
			err := txn.WithContext(ctx).
				Model(&app.InstallRoles{}).
				Where(app.InstallRoles{ID: existing.ID}).
				Update("app_role_config_id", role.ID).Error
			if err != nil {
				return fmt.Errorf("unable to repoint install role %s: %w", existing.ID, err)
			}
		}

		for _, stale := range staleByName {
			err := txn.WithContext(ctx).
				Where(app.InstallRoles{ID: stale.ID}).
				Delete(&app.InstallRoles{}).Error
			if err != nil {
				return fmt.Errorf("unable to delete install role %s: %w", stale.ID, err)
			}
		}
	}

	return nil
}
