package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

func (m *Migrations) Migration09NullWorkflowInstallID(ctx context.Context, db *gorm.DB) error {
	dropWorkflowStepsInstallID := "ALTER TABLE install_workflow_steps DROP COLUMN IF EXISTS install_id;"
	dropWorkflowsInstallID := "ALTER TABLE install_workflows DROP COLUMN IF EXISTS install_id;"
	if res := db.WithContext(ctx).
		Exec(dropWorkflowStepsInstallID); res.Error != nil {
		return res.Error
	}
	if res := db.WithContext(ctx).
		Exec(dropWorkflowsInstallID); res.Error != nil {
		return res.Error
	}
	return nil
}
