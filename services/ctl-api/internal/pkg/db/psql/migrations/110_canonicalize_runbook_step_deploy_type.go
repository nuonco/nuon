package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration110CanonicalizeRunbookStepDeployType(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).Exec(`
		UPDATE runbook_step_configs
		SET type = 'component_deploy'
		WHERE type = 'deploy'
		  AND deleted_at = 0;
	`); res.Error != nil {
		return res.Error
	}
	return nil
}
