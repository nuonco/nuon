package breakglass

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

// Sync owns its own AppBreakGlassConfig row rather than relying on the roles
// permissions.Sync attaches: every consumer reads AppConfig.BreakGlassConfig.
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.BreakGlass == nil {
		return nil
	}

	obj, err := build.BreakGlassConfig(appID, appConfigID, cfg.BreakGlass.Roles)
	if err != nil {
		return sync.SyncErr{
			Resource:    "break-glass",
			Description: err.Error(),
		}
	}

	if res := db.WithContext(ctx).Create(obj); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app break glass config",
			Err:         res.Error,
		}
	}

	return nil
}
