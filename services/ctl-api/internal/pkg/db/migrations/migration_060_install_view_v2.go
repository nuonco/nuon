package migrations

import (
	"context"
	_ "embed"
)

//go:embed installs_view_v2.sql
var installsViewV2 string

func (a *Migrations) migration060InstallsViewV2(ctx context.Context) error {
	if res := a.db.WithContext(ctx).Exec(installsViewV2); res.Error != nil {
		return res.Error
	}

	return nil
}
