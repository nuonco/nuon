package syncer

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// syncAppRunner creates the app runner configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_runner_config.go
func (s *syncer) syncAppRunner(ctx context.Context) error {
	// Convert env vars to pgtype.Hstore
	envVars := make(map[string]*string)
	for k, v := range s.cfg.Runner.EnvVarMap {
		val := v
		envVars[k] = &val
	}

	appRunnerConfig := app.AppRunnerConfig{
		AppConfigID:   s.appConfigID,
		AppID:         s.appID,
		HelmDriver:    app.AppRunnerConfigHelmDriverType(s.cfg.Runner.HelmDriver),
		EnvVars:       pgtype.Hstore(envVars),
		InitScriptURL: s.cfg.Runner.InitScriptURL,
		Type:          app.AppRunnerType(s.cfg.Runner.RunnerType),
	}

	res := s.db.WithContext(ctx).Create(&appRunnerConfig)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app runner config",
			Err:         res.Error,
		}
	}

	// Update the runner configs on all installs in the app
	res = s.db.WithContext(ctx).Model(&app.Install{}).
		Where("app_id = ?", s.appID).
		Update("app_runner_config_id", appRunnerConfig.ID)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to update app installs with new runner config",
			Err:         fmt.Errorf("unable to update app installs to reference new runner config: %w", res.Error),
		}
	}

	s.state.RunnerConfigID = appRunnerConfig.ID
	return nil
}
