package migrations

import (
	"context"
)

func (a *Migrations) migration067DropRunnerJobOwnerIndex(ctx context.Context) error {
	sql := `
DROP INDEX IF EXISTS idx_owner_name
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
