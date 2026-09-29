package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration098BackfillQueueSignalEnqueueFinishedAt(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec(`UPDATE queue_signals
SET enqueued = true,
    status = jsonb_set(
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
