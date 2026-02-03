package migrations

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"gorm.io/gorm"
)

func (m *Migrations) Migration091DeleteOrphanedActionTriggers(ctx context.Context, db *gorm.DB, cfg *internal.Config) error {
	query := `DELETE FROM action_workflow_trigger_configs
WHERE component_id IS NOT NULL
  AND component_id NOT IN (
    SELECT id FROM components WHERE id IS NOT NULL
  );
`

	if res := db.WithContext(ctx).
		Exec(query); res.Error != nil {
		return res.Error
	}

	return nil
}
