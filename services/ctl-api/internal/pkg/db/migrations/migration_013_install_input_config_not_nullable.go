package migrations

import "context"

func (a *Migrations) migration013InstallInputParentNotNull(ctx context.Context) error {
	sql := `
ALTER TABLE install_inputs ALTER COLUMN app_input_config_id SET NOT NULL;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
