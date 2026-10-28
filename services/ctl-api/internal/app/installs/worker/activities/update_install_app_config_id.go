package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/driftresolution"
)

type UpdateInstallAppConfigIDInput struct {
	InstallID      string `json:"install_id" validate:"required"`
	NewAppConfigID string `json:"new_app_config_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) UpdateInstallAppConfigID(ctx context.Context, input *UpdateInstallAppConfigIDInput) error {
	if input.InstallID == "" || input.NewAppConfigID == "" {
		return fmt.Errorf("install ID and app config ID are required")
	}
	var cfg app.AppConfig
	if err := a.db.WithContext(ctx).Where(app.AppConfig{ID: input.NewAppConfigID}).First(&cfg).Error; err != nil {
		return err
	}
	if cfg.BuildsPinned && !cfg.BuildsLocked {
		return fmt.Errorf("app config %s builds are not ready for rollout", cfg.ID)
	}
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := driftresolution.LockInstall(ctx, tx, input.InstallID); err != nil {
			return err
		}
		if err := tx.Model(&app.Install{}).Where(app.Install{ID: input.InstallID}).
			Updates(map[string]interface{}{
				"app_config_id": input.NewAppConfigID,
				"app_config_ref": gorm.Expr(
					"COALESCE(NULLIF(app_config_ref, 'null'::jsonb), '{}'::jsonb) || jsonb_build_object('expected_config_id', ?::text)",
					input.NewAppConfigID,
				),
			}).Error; err != nil {
			return fmt.Errorf("unable to update install app_config_id: %w", err)
		}
		return nil
	})
}
