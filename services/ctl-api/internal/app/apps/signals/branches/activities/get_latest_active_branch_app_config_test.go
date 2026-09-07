package activities

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupLatestActiveBranchAppConfigDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE installs (
			id TEXT PRIMARY KEY,
			app_id TEXT NOT NULL,
			app_config_id TEXT NOT NULL DEFAULT '',
			deleted_at INTEGER NOT NULL DEFAULT 0
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE app_configs (
			id TEXT PRIMARY KEY,
			app_id TEXT NOT NULL,
			app_branch_id TEXT,
			status TEXT NOT NULL,
			labels TEXT,
			created_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0
		)
	`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO installs (id, app_id) VALUES ('install-1', 'app-1')
	`).Error)

	return db
}

func TestGetLatestActiveBranchAppConfigExcludesPreview(t *testing.T) {
	db := setupLatestActiveBranchAppConfigDB(t)
	now := time.Now()
	require.NoError(t, db.Exec(`
		INSERT INTO app_configs (id, app_id, app_branch_id, status, labels, created_at)
		VALUES
			('regular-config', 'app-1', 'branch-1', 'active', '{}', ?),
			('preview-config', 'app-1', 'branch-1', 'active', ?, ?)
	`, now.Add(-time.Minute), `{"source":"git-preview-run"}`, now).Error)

	result, err := (&Activities{db: db}).GetLatestActiveBranchAppConfig(context.Background(), &GetLatestActiveBranchAppConfigInput{
		AppBranchID: "branch-1",
		InstallID:   "install-1",
	})
	require.NoError(t, err)
	require.Equal(t, "regular-config", result.AppConfigID)
	require.False(t, result.AlreadyCurrent)
}

func TestGetLatestActiveBranchAppConfigDoesNotFallbackToPreview(t *testing.T) {
	db := setupLatestActiveBranchAppConfigDB(t)
	require.NoError(t, db.Exec(`
		INSERT INTO app_configs (id, app_id, app_branch_id, status, labels, created_at)
		VALUES ('preview-config', 'app-1', 'branch-1', 'active', ?, ?)
	`, `{"source":"git-preview-run"}`, time.Now()).Error)

	result, err := (&Activities{db: db}).GetLatestActiveBranchAppConfig(context.Background(), &GetLatestActiveBranchAppConfigInput{
		AppBranchID: "branch-1",
		InstallID:   "install-1",
	})
	require.NoError(t, err)
	require.Empty(t, result.AppConfigID)
	require.False(t, result.AlreadyCurrent)
}
