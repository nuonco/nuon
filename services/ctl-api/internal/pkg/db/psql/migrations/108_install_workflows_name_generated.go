package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration108InstallWorkflowsNameGenerated adds the install_workflows.name
// STORED generated column that holds the human-readable workflow title.
// The expression itself lives in InstallWorkflowNameExpr — see
// install_workflow_name_expr.go — so future migrations that need to update
// the title (e.g. when a new workflow type is added) can edit the const and
// add a new migration that calls RebuildInstallWorkflowsNameColumnSQL.
func (m *Migrations) Migration108InstallWorkflowsNameGenerated(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(RebuildInstallWorkflowsNameColumnSQL()).Error
}
