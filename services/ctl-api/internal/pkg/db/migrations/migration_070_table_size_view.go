package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/table_sizes_v1.sql
var tableSizesViewV1 string

func (a *Migrations) migration070TableSizesView(ctx context.Context) error {
	if res := a.db.WithContext(ctx).Exec(tableSizesViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
