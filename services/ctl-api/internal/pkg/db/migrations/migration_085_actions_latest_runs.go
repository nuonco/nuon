package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/install_action_workflow_latest_runs_view_v1.sql
var iawrViewV1SQL string

func (a *Migrations) migration085ActionsLatestRuns(ctx context.Context) error {
	if res := a.db.WithContext(ctx).
		Exec(iawrViewV1SQL); res.Error != nil {
		return res.Error
	}

	return nil
}
