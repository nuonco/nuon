package migrations

import (
	"context"
	_ "embed"
)

func (a *Migrations) migration081DropRunnerHealtCheckIndex(ctx context.Context) error {
	// NOTE: secondary index is not needed because runner_id is already part of the primary key
	dropSQL := `DROP INDEX IF EXISTS idx_runner_health_checks_runner_id ON runner_health_checks ON CLUSTER simple`
	if res := a.chDB.WithContext(ctx).
		Exec(dropSQL); res.Error != nil {
		return res.Error
	}

	return nil
}
