package migrations

import (
	"context"
)

func (a *Migrations) migration071DropSettingsRefreshTimeout(ctx context.Context) error {
	sql := `
ALTER TABLE runner_group_settings DROP COLUMN IF EXISTS settings_refresh_timeout
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
