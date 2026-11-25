package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// The install's runner pointer follows its desired ACV: it is set at install
// creation from that ACV and repointed here when the install moves. An ACV
// without a runner config keeps the current pointer, matching install creation.
func (s *Helpers) MigrateInstallRunnerConfig(ctx context.Context, installID string) error {
	appCfg, err := s.installAppConfig(ctx, installID)
	if err != nil {
		return err
	}

	var runnerCfg app.AppRunnerConfig
	if err := s.db.WithContext(ctx).
		Select("id").
		Where(app.AppRunnerConfig{AppConfigID: appCfg.ID}).
		Order("created_at DESC").
		First(&runnerCfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get app runner config for install %s: %w", installID, err)
	}

	if err := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Where(app.Install{ID: installID}).
		Update("app_runner_config_id", runnerCfg.ID).Error; err != nil {
		return fmt.Errorf("unable to repoint install %s runner config: %w", installID, err)
	}

	return nil
}
