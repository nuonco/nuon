package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// sqlite cannot express Postgres labels->>'x' or char_length checks; these
// tests cover FindBaseAppBranchRun selection order and CreateAppBranchRun
// comparison wiring with a labels JSON column and hand-written DDL.

func setupAppBranchRunComparisonDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

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

func insertBranchRun(t *testing.T, db *gorm.DB, id, branchID, runType string, planOnly bool, buildsCompleted *bool, createdAt time.Time) {
	t.Helper()

	var labelsJSON any
	if buildsCompleted != nil {
		if *buildsCompleted {
			labelsJSON = `{"builds_completed":"true"}`
		} else {
			labelsJSON = `{"builds_completed":"false"}`
		}
	}

	planOnlyInt := 0
	if planOnly {
		planOnlyInt = 1
	}

	require.NoError(t, db.Exec(`
		INSERT INTO app_branch_runs (
			id, org_id, app_branch_id, app_branch_config_id, created_by_id,
			created_at, updated_at, status, run_type, plan_only, labels
		) VALUES (?, 'org-1', ?, 'cfg-1', 'acc-1', ?, ?, 'success', ?, ?, ?)
	`, id, branchID, createdAt, createdAt, runType, planOnlyInt, labelsJSON).Error)
}

func TestFindBaseAppBranchRun(t *testing.T) {
	now := time.Now()
	trueVal := true
	falseVal := false

	t.Run("returns most recent deploy with builds_completed true", func(t *testing.T) {
		db := setupAppBranchRunComparisonDB(t)
		insertBranchRun(t, db, "old", "branch-1", string(app.AppBranchRunTypeGit), false, &trueVal, now.Add(-2*time.Hour))
		insertBranchRun(t, db, "newer", "branch-1", string(app.AppBranchRunTypeManual), false, &trueVal, now.Add(-time.Hour))
		insertBranchRun(t, db, "preview", "branch-1", string(app.AppBranchRunTypeGitPreview), false, &trueVal, now)

		h := &Helpers{db: db}
		base, err := h.FindBaseAppBranchRun(context.Background(), "branch-1")
		require.NoError(t, err)
		require.Equal(t, "newer", base.ID)
	})

	t.Run("excludes preview and plan-only runs", func(t *testing.T) {
		db := setupAppBranchRunComparisonDB(t)
		insertBranchRun(t, db, "preview", "branch-1", string(app.AppBranchRunTypeGitPreview), false, &trueVal, now)
		insertBranchRun(t, db, "plan-only", "branch-1", string(app.AppBranchRunTypeManual), true, &trueVal, now.Add(-time.Minute))
		insertBranchRun(t, db, "deploy", "branch-1", string(app.AppBranchRunTypeGit), false, &trueVal, now.Add(-2*time.Hour))

		h := &Helpers{db: db}
		base, err := h.FindBaseAppBranchRun(context.Background(), "branch-1")
		require.NoError(t, err)
		require.Equal(t, "deploy", base.ID)
	})

	t.Run("excludes builds_completed false", func(t *testing.T) {
		db := setupAppBranchRunComparisonDB(t)
		insertBranchRun(t, db, "failed-builds", "branch-1", string(app.AppBranchRunTypeGit), false, &falseVal, now)
		insertBranchRun(t, db, "ok", "branch-1", string(app.AppBranchRunTypeGit), false, &trueVal, now.Add(-time.Hour))

		h := &Helpers{db: db}
		base, err := h.FindBaseAppBranchRun(context.Background(), "branch-1")
		require.NoError(t, err)
		require.Equal(t, "ok", base.ID)
	})

	t.Run("returns not found when no candidate exists", func(t *testing.T) {
		db := setupAppBranchRunComparisonDB(t)
		h := &Helpers{db: db}
		_, err := h.FindBaseAppBranchRun(context.Background(), "branch-1")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestCreateAppBranchRunCreatesComparison(t *testing.T) {
	now := time.Now()
	trueVal := true

	t.Run("links comparison to base deploy run", func(t *testing.T) {
		db := setupAppBranchRunComparisonDB(t)
		insertBranchRun(t, db, "base-run", "branch-1", string(app.AppBranchRunTypeGit), false, &trueVal, now.Add(-time.Hour))

		h := &Helpers{db: db}
		run, err := h.CreateAppBranchRun(context.Background(), &CreateAppBranchRunRequest{
			AppBranchID:       "branch-1",
			AppBranchConfigID: "cfg-1",
			RunType:           app.AppBranchRunTypeGit,
			Labels:            labels.Labels{"commit": "abc"},
		})
		require.NoError(t, err)
		require.NotNil(t, run.Comparison)
		require.Equal(t, run.ID, run.Comparison.HeadRunID)

		var comparison app.AppBranchRunComparison
		require.NoError(t, db.Where(app.AppBranchRunComparison{HeadRunID: run.ID}).First(&comparison).Error)
		require.Equal(t, run.ID, comparison.HeadRunID)
		require.NotNil(t, comparison.BaseRunID)
		require.Equal(t, "base-run", *comparison.BaseRunID)
	})

	t.Run("first run has nil base", func(t *testing.T) {
		db := setupAppBranchRunComparisonDB(t)
		h := &Helpers{db: db}
		run, err := h.CreateAppBranchRun(context.Background(), &CreateAppBranchRunRequest{
			AppBranchID:       "branch-1",
			AppBranchConfigID: "cfg-1",
			RunType:           app.AppBranchRunTypeGitPreview,
		})
		require.NoError(t, err)
		require.NotNil(t, run.Comparison)

		var comparison app.AppBranchRunComparison
		require.NoError(t, db.Where(app.AppBranchRunComparison{HeadRunID: run.ID}).First(&comparison).Error)
		require.Equal(t, run.ID, comparison.HeadRunID)
		require.Nil(t, comparison.BaseRunID)
	})

	t.Run("plan-only manual run skips comparison", func(t *testing.T) {
		db := setupAppBranchRunComparisonDB(t)
		h := &Helpers{db: db}
		run, err := h.CreateAppBranchRun(context.Background(), &CreateAppBranchRunRequest{
			AppBranchID:       "branch-1",
			AppBranchConfigID: "cfg-1",
			RunType:           app.AppBranchRunTypeManual,
			PlanOnly:          true,
		})
		require.NoError(t, err)
		require.Nil(t, run.Comparison)

		var count int64
		require.NoError(t, db.Model(&app.AppBranchRunComparison{}).Where(app.AppBranchRunComparison{HeadRunID: run.ID}).Count(&count).Error)
		require.Zero(t, count)
	})
}

func TestShouldCreateComparison(t *testing.T) {
	require.True(t, shouldCreateComparison(app.AppBranchRunTypeGit, false))
	require.True(t, shouldCreateComparison(app.AppBranchRunTypeGitPreview, false))
	require.True(t, shouldCreateComparison(app.AppBranchRunTypeManual, false))
	require.False(t, shouldCreateComparison(app.AppBranchRunTypeManual, true))
}
