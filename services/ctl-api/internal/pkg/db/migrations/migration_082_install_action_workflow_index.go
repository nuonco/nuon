package migrations

import (
	"context"
	_ "embed"
)

func (a *Migrations) migration082InstallActionWorkflowsIdx(ctx context.Context) error {
	dropSQL := `DROP INDEX IF EXISTS idx_install_action_workflow_id`
	if res := a.db.WithContext(ctx).
		Exec(dropSQL); res.Error != nil {
		return res.Error
	}

	createSQL := `CREATE UNIQUE INDEX IF NOT EXISTS idx_install_action_workflow_id
                ON install_action_workflows (deleted_at, install_id, action_workflow_id);`
	if res := a.db.WithContext(ctx).
		Exec(createSQL); res.Error != nil {
		return res.Error
	}

	return nil
}
