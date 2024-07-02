package migrations

import "context"

func (a *Migrations) migration043ComponentConfigConnectionsView(ctx context.Context) error {
	sql := `
  CREATE OR REPLACE VIEW component_config_connections_view_v1 AS
  SELECT *, row_number() OVER (PARTITION BY component_id
                               ORDER BY created_at) AS version
  FROM component_config_connections;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
