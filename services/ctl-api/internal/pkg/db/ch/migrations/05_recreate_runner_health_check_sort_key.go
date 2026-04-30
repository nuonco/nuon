package migrations

import (
	"context"
	_ "embed"
	"strings"

	"gorm.io/gorm"
)

//go:embed 05_recreate_runner_health_check_sort_key.sql
var RecreateRunnerHealthCheckSortKey string

func (m *Migrations) Migration005RecreateRunnerHealthCheckSortKey(ctx context.Context, db *gorm.DB) error {
	for _, stmt := range strings.Split(RecreateRunnerHealthCheckSortKey, ";") {
		stmt = stripSQLComments(stmt)
		if stmt == "" {
			continue
		}
		if res := db.WithContext(ctx).Exec(stmt); res.Error != nil {
			return res.Error
		}
	}

	return nil
}
