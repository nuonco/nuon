package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration110CanonicalizeRunbookStepDeployType rewrites the legacy 'deploy'
// runbook step type to its canonical name, 'component_deploy'. The rename was
// made to match 'component_tear_down'; the API and config layers still accept
// 'deploy' on input and canonicalize at ingress, so this only normalizes
// existing rows.
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
