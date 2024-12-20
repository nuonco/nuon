package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/app_configs_view_v2.sql
var appConfigsViewV1 string

func (a *Migrations) migration078AppConfigsViewV2(ctx context.Context) error {
	if res := a.db.WithContext(ctx).Exec(appConfigsViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
