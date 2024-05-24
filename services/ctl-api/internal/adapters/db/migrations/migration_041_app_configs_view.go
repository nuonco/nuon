package migrations

import "context"

func (a *Migrations) migration041AppConfigVersions(ctx context.Context) error {
	sql := `
  CREATE OR REPLACE VIEW app_configs_view AS
  SELECT *, row_number() OVER (PARTITION BY app_id
                               ORDER BY created_at) AS version
  FROM app_configs;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
