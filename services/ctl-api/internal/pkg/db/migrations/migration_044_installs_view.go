package migrations

import "context"

func (a *Migrations) migration044InstallsView(ctx context.Context) error {
	sql := `
  CREATE OR REPLACE VIEW installs_view_v1 AS
  SELECT *, 
  row_number() OVER (PARTITION BY app_id
                               ORDER BY created_at) AS install_number
  FROM installs;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
