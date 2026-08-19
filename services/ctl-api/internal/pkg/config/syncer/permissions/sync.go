package permissions

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

// Sync creates the app permissions configuration via the shared builder in
// internal/pkg/config/build, which the CreateAppPermissionsConfig handler also
// uses.
func Sync(ctx context.Context, db *gorm.DB, installHelpers *installhelpers.Helpers, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.Permissions == nil {
		return nil
	}

	var breakGlassRoles []*config.AppAWSIAMRole
	if cfg.BreakGlass != nil {
		breakGlassRoles = cfg.BreakGlass.Roles
	}

	obj, err := build.PermissionsConfig(build.PermissionsInput{
		AppID:           appID,
		AppConfigID:     appConfigID,
		Permissions:     cfg.Permissions,
		BreakGlassRoles: breakGlassRoles,
	})
	if err != nil {
		return sync.SyncErr{
			Resource:    "permissions",
			Description: err.Error(),
		}
	}

	// Repoint existing installs at the new role rows, or they keep resolving
	// against the previous permissions config.
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if res := tx.WithContext(ctx).Create(obj); res.Error != nil {
			return sync.SyncInternalErr{
				Description: "unable to create app permissions config",
				Err:         res.Error,
			}
		}
		if err := installHelpers.MigrateInstallRoles(ctx, tx, appID, *obj); err != nil {
			return sync.SyncInternalErr{
				Description: "unable to create app permissions config",
				Err:         err,
			}
		}

		return nil
	})

	return err
}
