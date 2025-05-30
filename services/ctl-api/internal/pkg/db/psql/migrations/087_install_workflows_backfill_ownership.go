package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

func (m *Migrations) Migration087InstallWorkflowsBackfillOwnership(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).
		Exec("UPDATE install_workflows SET owner_id = install_id, owner_type = 'installs' WHERE owner_id IS NULL"); res.Error != nil {
		return res.Error
	}
	if res := db.WithContext(ctx).
		Exec("UPDATE install_workflow_steps SET owner_id = install_id, owner_type = 'installs' WHERE owner_id IS NULL"); res.Error != nil {
		return res.Error
	}

	return nil
}
