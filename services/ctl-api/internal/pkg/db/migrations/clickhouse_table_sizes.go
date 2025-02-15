package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/clickhouse_table_sizes.sql
var clickhouseTableSizesViewV1 string

func (a *Migrations) migration083ClickhouseTableSizes(ctx context.Context) error {
	dropSQL := `DROP VIEW IF EXISTS table_sizes_view_v1 ON CLUSTER simple`
	if res := a.chDB.WithContext(ctx).
		Exec(dropSQL); res.Error != nil {
		return res.Error
	}

	if res := a.chDB.WithContext(ctx).
		Exec(clickhouseTableSizesViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
