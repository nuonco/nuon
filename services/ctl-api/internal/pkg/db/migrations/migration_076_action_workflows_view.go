package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/actions_workflow_view_v1.sql
var actionWorkflowsViewV1 string

func (a *Migrations) migration076ActionsWorkflowsView(ctx context.Context) error {
	if res := a.db.WithContext(ctx).Exec(actionWorkflowsViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
