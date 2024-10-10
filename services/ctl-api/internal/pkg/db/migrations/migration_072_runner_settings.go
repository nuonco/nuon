package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/runner_settings_v1.sql
var runnerSettingsViewV1 string

func (a *Migrations) migration072RunnerSettings(ctx context.Context) error {
	if res := a.db.WithContext(ctx).Exec(runnerSettingsViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
