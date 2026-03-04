package runner

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Sync creates the app runner configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_runner_config.go
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string, state *sync.State) error {
	// Convert env vars to pgtype.Hstore
	envVars := make(map[string]*string)
	for k, v := range cfg.Runner.EnvVarMap {
		val := v
		envVars[k] = &val
	}

	appRunnerConfig := app.AppRunnerConfig{
		AppConfigID:   appConfigID,
		AppID:         appID,
		HelmDriver:    app.AppRunnerConfigHelmDriverType(cfg.Runner.HelmDriver),
		EnvVars:       pgtype.Hstore(envVars),
		InitScriptURL: cfg.Runner.InitScriptURL,
		Type:          app.AppRunnerType(cfg.Runner.RunnerType),
	}

	res := db.WithContext(ctx).Create(&appRunnerConfig)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app runner config",
			Err:         res.Error,
		}
	}

	// Update the runner configs on all installs in the app
	res = db.WithContext(ctx).Model(&app.Install{}).
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
