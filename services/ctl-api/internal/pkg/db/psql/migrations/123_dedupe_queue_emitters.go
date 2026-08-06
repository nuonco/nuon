package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration123DedupeQueueEmitters deduplicates live emitters sharing a
// (queue_id, name) pair ahead of the idx_queue_emitters_live_uq partial unique
// index (declared in QueueEmitter.Indexes), keeping the newest — it carries the
// most recent config — and soft-deleting the older ones. The rn offset keeps
// the newly deleted rows' timestamps distinct. Also drops the abandoned
// full-tuple index an earlier iteration created via model tags.
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

	return db.WithContext(ctx).Exec(`
		DROP INDEX IF EXISTS idx_queue_emitters_queue_name
	`).Error
}
