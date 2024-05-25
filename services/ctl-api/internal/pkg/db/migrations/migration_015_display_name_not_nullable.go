package migrations

import "context"

func (a *Migrations) migration015DisplayNameNotNullable(ctx context.Context) error {
	sql := `
ALTER TABLE app_inputs ALTER COLUMN display_name SET NOT NULL;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
