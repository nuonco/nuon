package migrations

import "context"

func (a *Migrations) migration040ComponentConfigVersions(ctx context.Context) error {
	sql := `
UPDATE component_config_connections SET version=1 where version < 1;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
