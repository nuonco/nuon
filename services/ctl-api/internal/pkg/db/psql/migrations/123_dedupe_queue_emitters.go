package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration123DedupeQueueEmitters dedupes live emitters per (queue_id, name),
// keeping the newest, then creates the partial unique index. Creation lives
// here instead of QueueEmitter.Indexes because the indexes phase runs before
// custom migrations and would fail while duplicates exist.
func (m *Migrations) Migration123DedupeQueueEmitters(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Exec(`
		UPDATE queue_emitters qe
		SET deleted_at = EXTRACT(EPOCH FROM now())::bigint + d.rn
		FROM (
			SELECT id, ROW_NUMBER() OVER (PARTITION BY queue_id, name ORDER BY created_at DESC, id DESC) AS rn
			FROM queue_emitters
			WHERE deleted_at = 0
		) d
		WHERE qe.id = d.id AND d.rn > 1
	`).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(`
		DROP INDEX IF EXISTS idx_queue_emitters_queue_name
	`).Error; err != nil {
		return err
	}

	return db.WithContext(ctx).Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_queue_emitters_live_uq
		ON queue_emitters (queue_id, name) WHERE deleted_at = 0
	`).Error
}
