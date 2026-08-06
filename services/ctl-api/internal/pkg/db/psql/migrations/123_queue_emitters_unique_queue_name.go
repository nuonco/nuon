package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration123QueueEmittersUniqueQueueName deduplicates live emitters sharing a
// (queue_id, name) pair, then enforces uniqueness on live rows only so
// check-then-create ensure paths can't double-create emitters. Partial index:
// soft-deleted history may collide freely. Runs as a custom migration (not a
// model tag) because AutoMigrate executes before dedupe could.
func (m *Migrations) Migration123QueueEmittersUniqueQueueName(ctx context.Context, db *gorm.DB) error {
	// Keep the newest live emitter per (queue_id, name) — it carries the most
	// recent config — and soft-delete the older ones. The rn offset keeps the
	// newly deleted rows' timestamps distinct.
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

	// Earlier iterations created this as a full-tuple unique index via model
	// tags; drop it so soft-delete timestamp collisions can't fail updates.
	if err := db.WithContext(ctx).Exec(`
		DROP INDEX IF EXISTS idx_queue_emitters_queue_name
	`).Error; err != nil {
		return err
	}

	return db.WithContext(ctx).Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_queue_emitters_live_queue_name
		ON queue_emitters (queue_id, name) WHERE deleted_at = 0
	`).Error
}
