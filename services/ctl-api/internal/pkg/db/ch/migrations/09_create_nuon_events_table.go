package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

//go:embed 09_create_nuon_events_table.sql
var CreateNuonEventsTable string

func (m *Migrations) Migration009CreateNuonEventsTable(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).Exec(CreateNuonEventsTable); res.Error != nil {
		return res.Error
	}
	return nil
}
