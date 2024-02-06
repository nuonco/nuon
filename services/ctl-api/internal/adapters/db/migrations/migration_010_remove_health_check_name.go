package migrations

import (
	"context"
)

func (a *Migrations) migration010RemoveHealthCheckName(ctx context.Context) error {
	sql := `
ALTER TABLE  org_health_checks DROP COLUMN IF EXISTS name;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}
	return nil
}
