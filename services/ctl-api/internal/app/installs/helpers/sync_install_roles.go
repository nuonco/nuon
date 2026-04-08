package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// SyncInstallRoles synchronises InstallRoles for every install belonging to
// the given app so they reflect the roles in the current AppPermissionsConfig.
// New roles are created; roles that no longer exist in the config are soft-deleted.
func (s *Helpers) SyncInstallRoles(ctx context.Context, appID string, permCfg app.AppPermissionsConfig) error {
	if permCfg.ID == "" {
		return nil
	}

	var installs []app.Install
	res := s.db.WithContext(ctx).
		Preload("InstallRoles").
		Where("app_id = ?", appID).
		Find(&installs)
	if res.Error != nil {
		return fmt.Errorf("unable to get installs for app %s: %w", appID, res.Error)
	}

	// build a set of desired role config IDs
	desiredRoles := make(map[string]struct{}, len(permCfg.Roles))
	for _, role := range permCfg.Roles {
		desiredRoles[role.ID] = struct{}{}
	}

	for _, install := range installs {
		// build set of existing role config IDs for this install
		existingRoles := make(map[string]string, len(install.InstallRoles)) // appRoleConfigID -> installRoleID
		for _, ir := range install.InstallRoles {
			existingRoles[ir.AppRoleConfigID] = ir.ID
		}

		// add new roles
		for _, role := range permCfg.Roles {
			if _, exists := existingRoles[role.ID]; exists {
				continue
			}
			newRole := app.InstallRoles{
				InstallID:       install.ID,
				AppRoleConfigID: role.ID,
			}
			if err := s.db.WithContext(ctx).Create(&newRole).Error; err != nil {
				return fmt.Errorf("unable to create install role for install %s: %w", install.ID, err)
			}
		}

		// soft-delete removed roles
		for appRoleConfigID, installRoleID := range existingRoles {
			if _, desired := desiredRoles[appRoleConfigID]; desired {
				continue
			}
			if err := s.db.WithContext(ctx).Delete(&app.InstallRoles{}, "id = ?", installRoleID).Error; err != nil {
				return fmt.Errorf("unable to delete install role %s: %w", installRoleID, err)
			}
		}
	}

	return nil
}
