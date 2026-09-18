package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration133DedupeQueues(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Exec(`
		UPDATE queues SET name = '' WHERE name IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(`
		WITH ranked AS (
			SELECT
				q.id,
				ROW_NUMBER() OVER (
					PARTITION BY q.owner_id, q.owner_type, q.name
					ORDER BY MAX(s.created_at) DESC NULLS LAST, q.created_at DESC, q.id DESC
				) AS rn
			FROM queues q
			LEFT JOIN queue_signals s ON s.queue_id = q.id AND s.deleted_at = 0
			WHERE q.deleted_at = 0
			GROUP BY q.id, q.owner_id, q.owner_type, q.name, q.created_at
		)
		UPDATE queues q
		SET deleted_at = EXTRACT(EPOCH FROM now())::bigint + ranked.rn
		FROM ranked
		WHERE q.id = ranked.id AND ranked.rn > 1
	`).Error; err != nil {
		return err
	}

	return db.WithContext(ctx).Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_queues_live_owner_name_uq
		ON queues (owner_id, owner_type, name) WHERE deleted_at = 0
	`).Error
}
