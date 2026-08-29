package activities

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCreateInstallGroupRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE orgs (id TEXT PRIMARY KEY, created_by_id TEXT NOT NULL DEFAULT '', created_at DATETIME, updated_at DATETIME, deleted_at INTEGER NOT NULL DEFAULT 0)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE accounts (id TEXT PRIMARY KEY, created_by_id TEXT NOT NULL DEFAULT '', created_at DATETIME, updated_at DATETIME, deleted_at INTEGER NOT NULL DEFAULT 0)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE app_branch_runs (
			id TEXT PRIMARY KEY, org_id TEXT NOT NULL, created_by_id TEXT NOT NULL,
			created_at DATETIME, updated_at DATETIME, deleted_at INTEGER NOT NULL DEFAULT 0,
			app_branch_id TEXT NOT NULL DEFAULT '', app_branch_config_id TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pending', run_type TEXT NOT NULL DEFAULT 'manual-run',
			plan_only INTEGER NOT NULL DEFAULT 0, force INTEGER NOT NULL DEFAULT 0
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE app_branch_install_groups (
			id TEXT PRIMARY KEY, org_id TEXT NOT NULL, created_by_id TEXT NOT NULL,
			created_at DATETIME, updated_at DATETIME, deleted_at INTEGER NOT NULL DEFAULT 0,
			app_branch_config_id TEXT NOT NULL DEFAULT '', name TEXT NOT NULL DEFAULT ''
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE install_group_runs (
			id TEXT PRIMARY KEY, org_id TEXT NOT NULL, created_by_id TEXT NOT NULL,
			created_at DATETIME, updated_at DATETIME, deleted_at INTEGER NOT NULL DEFAULT 0,
			app_branch_run_id TEXT NOT NULL, install_group_id TEXT NOT NULL,
			install_group_name TEXT NOT NULL, status TEXT, total_installs INTEGER
		)
	`).Error)
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, created_at, updated_at) VALUES ('org-1', datetime('now'), datetime('now'))`).Error)
	require.NoError(t, db.Exec(`INSERT INTO accounts (id, created_at, updated_at) VALUES ('acc-1', datetime('now'), datetime('now'))`).Error)
	now := time.Now()
	require.NoError(t, db.Exec(`
		INSERT INTO app_branch_runs (id, org_id, created_by_id, created_at, updated_at, app_branch_id, app_branch_config_id)
		VALUES ('run-1', 'org-1', 'acc-1', ?, ?, 'branch-1', 'cfg-1')
	`, now, now).Error)

	return db
}

func TestCreateInstallGroupRunEmptyInstallGroupIDFails(t *testing.T) {
	db := setupCreateInstallGroupRunDB(t)
	a := &Activities{db: db}

	_, err := a.CreateInstallGroupRun(context.Background(), &CreateInstallGroupRunInput{
		AppBranchRunID:   "run-1",
		InstallGroupID:   "",
		InstallGroupName: "preview",
		TotalInstalls:    1,
	})
	require.Error(t, err)
}
