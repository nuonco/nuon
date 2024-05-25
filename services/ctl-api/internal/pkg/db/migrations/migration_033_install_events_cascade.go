package migrations

import (
	"context"
)

func (a *Migrations) migration033InstallEventsCascade(ctx context.Context) error {
	sql := `
ALTER TABLE install_events DROP CONSTRAINT IF EXISTS fk_install_events_install;
ALTER TABLE install_events ADD CONSTRAINT fk_install_events_install
	FOREIGN KEY (install_id)
	REFERENCES installs(id)
	ON DELETE CASCADE;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
