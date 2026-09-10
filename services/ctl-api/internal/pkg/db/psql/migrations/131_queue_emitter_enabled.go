package migrations

import (
	"context"

	"gorm.io/gorm"
)

// QueueEmitter is not in psql.AllModels() — the table only exists because
// AutoMigrate creates it while wiring the Queue.Emitters foreign key, and that
// path does not add columns to an existing table. Every column the model gains
// has to be added here.
//
// The NOT NULL DEFAULT true backfills existing rows to enabled, which is the
// state they were implicitly in before the column existed.
func (m *Migrations) Migration131QueueEmitterEnabled(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
		ALTER TABLE queue_emitters ADD COLUMN IF NOT EXISTS enabled boolean NOT NULL DEFAULT true;
		ALTER TABLE queue_emitters ADD COLUMN IF NOT EXISTS disabled_reason text;
	`).Error
}
