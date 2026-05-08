package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

//go:embed 07_create_queries_table.sql
var CreateQueriesTable string

func (m *Migrations) Migration007CreateQueriesTable(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).Exec(CreateQueriesTable); res.Error != nil {
		return res.Error
	}
	return nil
}
