package migrations

import (
	"context"
)

func (a *Migrations) migration034AppSandboxConfigAppID(ctx context.Context) error {
	// delete all installs that are connected to an app sandbox config that does not have an app id
	sql := `
	DELETE FROM installs WHERE id IN
		(SELECT installs.id FROM installs
			JOIN app_sandbox_configs ON app_sandbox_configs.id=installs.app_sandbox_config_id
			WHERE app_sandbox_configs.app_id is null);
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	// delete all sandbox configs that do not have an app id
	sql = `
	DELETE FROM app_sandbox_configs WHERE app_id IS NULL;
	`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
