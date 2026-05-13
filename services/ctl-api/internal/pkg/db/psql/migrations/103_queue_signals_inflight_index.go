package migrations

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Replaces a manually-created sibling whose partial predicate referenced
// 'in_progress' instead of the Go enum value 'in-progress', so it was never used.
func (m *Migrations) Migration103QueueSignalsInflightIndex(ctx context.Context, db *gorm.DB) error {
	const name = "idx_queue_signals_emitter_id_queue_id_inflight"

	var indexdef string
	if err := db.WithContext(ctx).
		Raw(`SELECT indexdef FROM pg_indexes WHERE indexname = ?`, name).
		Scan(&indexdef).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if strings.Contains(indexdef, "'in_progress'") {
		if res := db.WithContext(ctx).Exec(
			`DROP INDEX CONCURRENTLY IF EXISTS ` + name,
		); res.Error != nil {
			return res.Error
		}
	}

	res := db.WithContext(ctx).Exec(
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS ` + name + `
ON queue_signals (emitter_id, queue_id)
WHERE deleted_at = 0
  AND (status->>'status') IN ('queued', 'in-progress')`,
	)
	return res.Error
}
