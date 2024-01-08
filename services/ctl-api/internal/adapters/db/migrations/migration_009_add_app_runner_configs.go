package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration009AddAppRunnerConfigs(ctx context.Context) error {
	var apps []*app.App
	res := a.db.WithContext(ctx).
		Preload("AppRunnerConfigs").
		Find(&apps)
	if res.Error != nil {
		return res.Error
	}

	for _, currentApp := range apps {
		if len(currentApp.AppRunnerConfigs) > 0 {
			continue
		}

		appRunnerCfg := app.AppRunnerConfig{
			OrgID:   currentApp.OrgID,
			AppID:   currentApp.ID,
			EnvVars: pgtype.Hstore(map[string]*string{}),
			Type:    app.AppRunnerTypeAWSEKS,
		}

		if res := a.db.WithContext(ctx).
			Create(&appRunnerCfg); res.Error != nil {
			return fmt.Errorf("unable to create app runner config: %w", res.Error)
		}

		// update the runner configs on all installs in the app
		res = a.db.WithContext(ctx).Model(&app.Install{}).
			Where("app_id = ?", currentApp.ID).
			Update("app_runner_config_id", appRunnerCfg.ID)
		if res.Error != nil {
			return fmt.Errorf("unable to update app installs to reference new runner config: %w", res.Error)
		}
	}

	return nil
}
