package migrations

import (
	"context"
)

func (a *Migrations) migration038DropAppInstallers(ctx context.Context) error {
	sql := `
DROP TABLE app_installers CASCADE;
DROP TABLE app_installer_metadata CASCADE;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
