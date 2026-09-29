package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration109BackfillRunbookStepDeployDependents(ctx context.Context, db *gorm.DB) error {
	var hasLegacy bool
	if err := db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'runbook_step_configs' AND column_name = 'deploy_dependencies'
		)
	`).Scan(&hasLegacy).Error; err != nil {
		return err
	}
	if !hasLegacy {
		return nil
	}

	var hasNew bool
	if err := db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'runbook_step_configs' AND column_name = 'deploy_dependents'
		)
	`).Scan(&hasNew).Error; err != nil {
		return err
	}
	if !hasNew {
		return nil
	}

	if res := db.WithContext(ctx).Exec(`
		UPDATE runbook_step_configs
		SET deploy_dependents = true
		WHERE deploy_dependencies = true
		  AND deploy_dependents = false
		  AND deleted_at = 0;
	`); res.Error != nil {
		return res.Error
	}

	return nil
}
