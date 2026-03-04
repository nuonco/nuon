package breakglass

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
)

// Sync creates the app break-glass configuration.
// Note: Break-glass roles are handled as part of permissions sync
// This is a no-op since break-glass roles are added to AppPermissionsConfig
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	// Break-glass roles are handled in permissions.Sync
	// This method exists to maintain compatibility with the sync steps
	return nil
}
