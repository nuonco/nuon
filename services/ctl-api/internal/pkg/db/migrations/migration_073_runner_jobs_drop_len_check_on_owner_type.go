package migrations

import (
	"context"
)

func (a *Migrations) migration073DropLengthCheckOnOwnerType(ctx context.Context) error {
	sql := `
ALTER TABLE runner_jobs DROP CONSTRAINT IF EXISTS owner_type_checker;
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
