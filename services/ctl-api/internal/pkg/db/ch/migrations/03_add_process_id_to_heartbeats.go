package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

//go:embed 03_add_process_id_to_heartbeats.sql
var AddProcessIDToHeartBeats string

func (m *Migrations) Migration003AddProcessIDToHeartBeats(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(AddProcessIDToHeartBeats); res.Error != nil {
		return res.Error
	}

	return nil
}
