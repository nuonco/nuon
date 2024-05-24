package migrations

import "context"

func (a *Migrations) migration042ComponentConfigConnectionsDropVersion(ctx context.Context) error {
	sql := `
ALTER TABLE component_config_connections DROP IF EXISTS version;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
