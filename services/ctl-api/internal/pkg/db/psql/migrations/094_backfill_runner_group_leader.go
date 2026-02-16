package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration094BackfillRunnerGroupLeader(ctx context.Context, db *gorm.DB) error {
	backfill := `UPDATE runner_groups rg
SET leader_runner_id = (
  SELECT r.id FROM runners r
  WHERE r.runner_group_id = rg.id
    AND r.status = 'active'
    AND r.deleted_at = 0
  ORDER BY r.created_at ASC
  LIMIT 1
)
WHERE rg.deleted_at = 0;`

	if res := db.WithContext(ctx).Exec(backfill); res.Error != nil {
		return res.Error
	}

	return nil
}
