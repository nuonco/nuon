package migrations

import (
	"context"
)

func (a *Migrations) migration039DropComponentVarNameRequired(ctx context.Context) error {
	sql := `
ALTER TABLE components ALTER COLUMN var_name DROP NOT NULL;
ALTER TABLE components ALTER COLUMN var_name SET DEFAULT NULL;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
