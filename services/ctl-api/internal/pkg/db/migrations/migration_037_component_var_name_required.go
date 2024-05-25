package migrations

import (
	"context"
)

func (a *Migrations) migration037ComponentVarNameRequired(ctx context.Context) error {
	sql := `
ALTER TABLE components ALTER COLUMN var_name SET NOT NULL;
ALTER TABLE components ALTER COLUMN var_name SET DEFAULT NULL;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
