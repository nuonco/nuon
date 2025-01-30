package migrations

import "context"

func (a *Migrations) migration079ActionRunsActiveToFinished(ctx context.Context) error {
	sql := `
  UPDATE install_action_workflow_runs
	SET status = 'finished'
	WHERE status = 'active';
`
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
