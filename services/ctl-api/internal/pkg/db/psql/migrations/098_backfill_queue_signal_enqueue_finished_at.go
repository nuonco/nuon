package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration098BackfillQueueSignalEnqueueFinishedAt sets enqueue_finished_at on
// all existing queue signals that don't have it. This ensures the new metric
// (queue_signals.missing_enqueue_finished) starts from a clean baseline when
// the background-enqueue change rolls out.
func (m *Migrations) Migration098BackfillQueueSignalEnqueueFinishedAt(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(`UPDATE queue_signals
SET status = jsonb_set(
    COALESCE(status::jsonb, '{}'::jsonb),
    '{metadata,enqueue_finished_at}',
    to_jsonb(to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'))
)
WHERE deleted_at = 0
AND NOT (COALESCE(status::jsonb, '{}'::jsonb) -> 'metadata' ? 'enqueue_finished_at');`); res.Error != nil {
		return res.Error
	}

	return nil
}
