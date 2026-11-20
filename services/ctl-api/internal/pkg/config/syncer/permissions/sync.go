package permissions

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

// Sync creates the app permissions configuration via the shared builder in
// internal/pkg/config/build, which the CreateAppPermissionsConfig handler also
// uses.
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
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
		StackType:       stackType(cfg),
	})
	if err != nil {
		return sync.SyncErr{
			Resource:    "permissions",
			Description: err.Error(),
			Err:         err,
		}
	}

	if res := db.WithContext(ctx).Create(obj); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app permissions config",
			Err:         res.Error,
		}
	}

	return nil
}

func stackType(cfg *config.AppConfig) string {
	if cfg == nil || cfg.Stack == nil {
		return ""
	}
	return cfg.Stack.Type
}
