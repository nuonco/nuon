package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/runner_jobs_view_v1.sql
var runnerJobsViewV1 string

func (a *Migrations) migration077RunnerJobsView(ctx context.Context) error {
	if res := a.db.WithContext(ctx).Exec(runnerJobsViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
