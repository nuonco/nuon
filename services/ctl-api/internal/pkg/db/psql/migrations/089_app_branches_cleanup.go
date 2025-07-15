package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

func (m *Migrations) Migration089AppBracnhesCleanup(ctx context.Context, db *gorm.DB) error {
	dropTableVCSConnectionBranches := `DROP TABLE IF EXISTS vcs_connection_branches CASCADE`
	dropTableVCSConnectionRepos := `DROP TABLE IF EXISTS vcs_connection_repos CASCADE`
	dropTableAppConfigSyncRuns := `DROP TABLE IF EXISTS app_config_sync_runs CASCADE`
	dropTableComponentPushes := `DROP TABLE IF EXISTS component_pushes CASCADE`
	dropTableComponentBuildConnections := `DROP TABLE IF EXISTS component_build_connections CASCADE`

	if res := db.WithContext(ctx).
		Exec(dropTableComponentBuildConnections); res.Error != nil {
		return res.Error
	}

	if res := db.WithContext(ctx).
		Exec(dropTableVCSConnectionBranches); res.Error != nil {
		return res.Error
	}

	if res := db.WithContext(ctx).
		Exec(dropTableAppConfigSyncRuns); res.Error != nil {
		return res.Error
	}

	if res := db.WithContext(ctx).
		Exec(dropTableComponentPushes); res.Error != nil {
		return res.Error
	}

	if res := db.WithContext(ctx).
		Exec(dropTableVCSConnectionRepos); res.Error != nil {
		return res.Error
	}

	return nil
}
