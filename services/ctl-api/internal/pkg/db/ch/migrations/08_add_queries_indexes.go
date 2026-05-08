package migrations

import (
	"context"

	"gorm.io/gorm"
)

var queriesIndexStatements = []string{
	"ALTER TABLE queries ON CLUSTER simple ADD INDEX IF NOT EXISTS idx_error error TYPE set(0) GRANULARITY 1",
	"ALTER TABLE queries ON CLUSTER simple ADD INDEX IF NOT EXISTS idx_source source TYPE set(0) GRANULARITY 1",
	"ALTER TABLE queries ON CLUSTER simple ADD INDEX IF NOT EXISTS idx_db_type db_type TYPE set(0) GRANULARITY 1",
	"ALTER TABLE queries ON CLUSTER simple ADD INDEX IF NOT EXISTS idx_endpoint endpoint TYPE set(0) GRANULARITY 1",
	"ALTER TABLE queries ON CLUSTER simple ADD INDEX IF NOT EXISTS idx_timestamp timestamp TYPE minmax GRANULARITY 1",
}

func (m *Migrations) Migration008AddQueriesIndexes(ctx context.Context, db *gorm.DB) error {
	for _, stmt := range queriesIndexStatements {
		if res := db.WithContext(ctx).Exec(stmt); res.Error != nil {
			return res.Error
		}
	}
	return nil
}
