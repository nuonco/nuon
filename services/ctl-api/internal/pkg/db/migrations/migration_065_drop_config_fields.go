package migrations

import "context"

func (a *Migrations) migration065ConfigFields(ctx context.Context) error {
	sql := `
ALTER TABLE app_configs DROP COLUMN IF EXISTS generated_terraform_json CASCADE;
ALTER TABLE app_configs DROP COLUMN IF EXISTS content CASCADE;
ALTER TABLE app_configs DROP COLUMN IF EXISTS format CASCADE;
`

	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	// Need to create the view again, after removing columns
	sql = `
  CREATE OR REPLACE VIEW app_configs_view_v1 AS
  SELECT *, row_number() OVER (PARTITION BY app_id
                               ORDER BY created_at) AS version
  FROM app_configs;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
