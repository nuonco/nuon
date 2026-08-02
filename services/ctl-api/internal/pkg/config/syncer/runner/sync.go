package runner

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

// Sync creates the app runner configuration via the shared builder in
// internal/pkg/config/build, which the CreateAppRunnerConfig handler also uses.
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string, state *sync.State) error {
	appRunnerConfig := build.RunnerConfig(build.RunnerInputFromConfig(*cfg.Runner, appID, appConfigID))

	if res := db.WithContext(ctx).Create(appRunnerConfig); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app runner config",
			Err:         res.Error,
		}
	}

	// Point every install in the app at the new runner config.
	res := db.WithContext(ctx).Model(&app.Install{}).
		Where("app_id = ?", appID).
		Update("app_runner_config_id", appRunnerConfig.ID)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to update app installs with new runner config",
			Err:         fmt.Errorf("unable to update app installs to reference new runner config: %w", res.Error),
		}
	}

	state.RunnerConfigID = appRunnerConfig.ID
	return nil
}
