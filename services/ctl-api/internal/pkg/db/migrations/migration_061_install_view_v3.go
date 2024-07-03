package migrations

import (
	"context"
	_ "embed"
)

//go:embed installs_view_v3.sql
var installsViewV3 string

func (a *Migrations) migration061InstallsViewV3(ctx context.Context) error {
	if res := a.db.WithContext(ctx).Exec(installsViewV3); res.Error != nil {
		return res.Error
	}

	return nil
}
