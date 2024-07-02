package migrations

import "context"

func (a *Migrations) migration059InstallStatus(ctx context.Context) error {
	sql := `
ALTER TABLE installs DROP COLUMN IF EXISTS status CASCADE;
ALTER TABLE installs DROP COLUMN IF EXISTS status_description CASCADE;
        `
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
