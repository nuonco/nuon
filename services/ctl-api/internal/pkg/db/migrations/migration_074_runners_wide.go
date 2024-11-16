package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/runner_wide_v1.sql
var runnerWideViewV1 string

func (a *Migrations) migration074RunnerWideView(ctx context.Context) error {
	if res := a.db.WithContext(ctx).Exec(runnerWideViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
