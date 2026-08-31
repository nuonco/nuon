package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func setupAppBranchRunPreviewDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE app_branches (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL DEFAULT '',
			app_id TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			created_by_id TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			managed_by TEXT NOT NULL DEFAULT 'manually'
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE app_branch_configs (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL DEFAULT '',
			app_branch_id TEXT NOT NULL,
			created_by_id TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			preview_config TEXT
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE installs (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL DEFAULT '',
			app_id TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			app_branch_id TEXT,
			deleted_at INTEGER NOT NULL DEFAULT 0
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE app_branch_runs (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL DEFAULT '',
			app_branch_id TEXT NOT NULL,
			app_branch_config_id TEXT NOT NULL DEFAULT '',
			created_by_id TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			run_type TEXT NOT NULL DEFAULT 'manual-run',
			plan_only INTEGER NOT NULL DEFAULT 0,
			force INTEGER NOT NULL DEFAULT 0,
			labels TEXT,
			app_config_id TEXT NOT NULL DEFAULT '',
			head_sha TEXT NOT NULL DEFAULT '',
			base_branch TEXT NOT NULL DEFAULT '',
			event_type TEXT NOT NULL DEFAULT '',
			no_config_changes INTEGER NOT NULL DEFAULT 0,
			error_message TEXT NOT NULL DEFAULT '',
			composite_error TEXT,
			trigger_event_dispatch_id TEXT,
			workflow_id TEXT,
			pr_number INTEGER,
			github_comment_id INTEGER,
			started_at DATETIME,
			completed_at DATETIME,
			log_stream_id TEXT,
			vcs_connection_commit_id TEXT
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE app_branch_run_previews (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL DEFAULT '',
			created_by_id TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			app_branch_run_id TEXT NOT NULL UNIQUE,
			source TEXT NOT NULL,
			mode TEXT NOT NULL,
			install_id TEXT,
			install_name TEXT,
			git_ref TEXT,
			input_app_config_id TEXT,
			branch_preview_config TEXT NOT NULL DEFAULT '{}',
			override_preview_config TEXT,
			resolved_preview_config TEXT NOT NULL DEFAULT '{}'
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE app_branch_run_comparisons (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL DEFAULT '',
			created_by_id TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			head_run_id TEXT NOT NULL,
			base_run_id TEXT,
			git_diff TEXT,
			full_diff TEXT,
			config_diff TEXT
		)
	`).Error)

	return db
}

func TestCreateAppBranchRunCreatesPreview(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupAppBranchRunPreviewDB(t)
	now := time.Now()

	require.NoError(t, db.Exec(`
		INSERT INTO app_branches (id, app_id, name, created_at, updated_at)
		VALUES ('branch-1', 'app-1', 'main', ?, ?)
	`, now, now).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO app_branch_configs (id, app_branch_id, created_at, updated_at)
		VALUES ('cfg-1', 'branch-1', ?, ?)
	`, now, now).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO installs (id, app_id, name)
		VALUES ('inst-b', 'app-1', 'beta')
	`).Error)

	prNumber := 32
	installID := "inst-b"
	h := &Helpers{db: db}
	run, err := h.CreateAppBranchRun(ctx, &CreateAppBranchRunRequest{
		AppBranchID:       "branch-1",
		AppBranchConfigID: "cfg-1",
		RunType:           app.AppBranchRunTypeGitPreview,
		Preview: &PreviewRunInput{
			Source:   app.AppBranchRunPreviewSourcePR,
			PRNumber: &prNumber,
			Override: &app.AppBranchPreviewOverride{
				InstallID: &installID,
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, run.Preview)
	require.True(t, run.PlanOnly)

	var count int64
	require.NoError(t, db.Model(&app.AppBranchRunPreview{}).Count(&count).Error)
	require.Equal(t, int64(1), count)

	var preview app.AppBranchRunPreview
	require.NoError(t, db.Where(app.AppBranchRunPreview{AppBranchRunID: run.ID}).First(&preview).Error)
	require.Equal(t, run.Preview.ID, preview.ID)
	require.Equal(t, installID, preview.InstallID)
}

func TestCreateAppBranchRunPreviewCreatesComparison(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupAppBranchRunPreviewDB(t)
	now := time.Now()

	require.NoError(t, db.Exec(`
		INSERT INTO app_branches (id, app_id, name, created_at, updated_at)
		VALUES ('branch-1', 'app-1', 'main', ?, ?)
	`, now, now).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO app_branch_configs (id, app_branch_id, created_at, updated_at)
		VALUES ('cfg-1', 'branch-1', ?, ?)
	`, now, now).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO installs (id, app_id, name)
		VALUES ('inst-b', 'app-1', 'beta')
	`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO app_branch_runs (
			id, app_branch_id, app_branch_config_id, run_type, plan_only, status, labels, created_at, updated_at
		) VALUES (
			'base-run', 'branch-1', 'cfg-1', 'manual-run', 0, 'success', '{"builds_completed":"true"}', ?, ?
		)
	`, now.Add(-time.Hour), now.Add(-time.Hour)).Error)

	prNumber := 32
	installID := "inst-b"
	h := &Helpers{db: db}
	run, err := h.CreateAppBranchRun(ctx, &CreateAppBranchRunRequest{
		AppBranchID:       "branch-1",
		AppBranchConfigID: "cfg-1",
		RunType:           app.AppBranchRunTypeGitPreview,
		Preview: &PreviewRunInput{
			Source:   app.AppBranchRunPreviewSourcePR,
			PRNumber: &prNumber,
			Override: &app.AppBranchPreviewOverride{
				InstallID: &installID,
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, run.Preview)
	require.NotNil(t, run.Comparison)
	require.Equal(t, run.ID, run.Comparison.HeadRunID)
	require.NotNil(t, run.Comparison.BaseRunID)
	require.Equal(t, "base-run", *run.Comparison.BaseRunID)
}
