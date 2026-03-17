package appconfig

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Sync updates the AppConfig record with metadata from the parsed config,
// including the readme and the config schema version.
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appConfigID string) error {
	res := db.WithContext(ctx).
		Model(&app.AppConfig{}).
		Where("id = ?", appConfigID).
		Select("readme", "cli_version").
		Updates(app.AppConfig{
			Readme:     cfg.Readme,
			CLIVersion: cfg.Version,
		})
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to update app config metadata",
			Err:         fmt.Errorf("unable to update app config metadata: %w", res.Error),
		}
	}

	return nil
}
