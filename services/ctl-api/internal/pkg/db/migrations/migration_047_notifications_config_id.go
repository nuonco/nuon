package migrations

import "context"

func (a *Migrations) migration047NotificationsConfig(ctx context.Context) error {
	sql := `
ALTER TABLE apps DROP COLUMN IF EXISTS notifications_config_id;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
