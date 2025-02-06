package migrations

import (
	"context"
	_ "embed"
)

//go:embed views/runner_health_checks_view_v1.sql
var runnerHealthChecksViewV1 string

func (a *Migrations) migration080RunnerHealthChecks(ctx context.Context) error {
	dropSQL := `DROP VIEW IF EXISTS runner_health_checks_v1`
	if res := a.chDB.WithContext(ctx).
		Exec(dropSQL); res.Error != nil {
		return res.Error
	}

	if res := a.chDB.WithContext(ctx).
		Exec(runnerHealthChecksViewV1); res.Error != nil {
		return res.Error
	}

	return nil
}
