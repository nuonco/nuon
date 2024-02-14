package migrations

import "context"

func (a *Migrations) migration011RemoveAppInputConfig(ctx context.Context) error {
	sql := `
ALTER TABLE installs DROP COLUMN IF EXISTS app_input_config;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
