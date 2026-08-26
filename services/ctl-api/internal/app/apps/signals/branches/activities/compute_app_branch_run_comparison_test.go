package activities

import (
	"context"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func setupComparisonActivityDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

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
	require.NoError(t, db.Exec(`
		CREATE TABLE app_branch_runs (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL DEFAULT '',
			app_branch_id TEXT NOT NULL DEFAULT '',
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
			vcs_connection_commit_id TEXT
		)
	`).Error)
	return db
}

func TestComputeAndStoreAppBranchRunComparison_NoComparisonRow(t *testing.T) {
	db := setupComparisonActivityDB(t)
	a := &Activities{
		v:  validator.New(),
		db: db,
		l:  zap.NewNop(),
	}

	out, err := a.ComputeAndStoreAppBranchRunComparison(context.Background(), &ComputeAndStoreAppBranchRunComparisonInput{
		AppBranchID: "appbranch0000000000000001",
		RunID:       "appbranchrun000000000001",
	})
	require.NoError(t, err)
	require.True(t, out.Skipped)
	require.Equal(t, "no comparison row", out.SkipReason)
}

func TestComputeAndStoreAppBranchRunComparison_NoBaseRun(t *testing.T) {
	db := setupComparisonActivityDB(t)
	now := time.Now()
	require.NoError(t, db.Exec(`
		INSERT INTO app_branch_run_comparisons
		(id, org_id, created_by_id, created_at, updated_at, head_run_id, base_run_id)
		VALUES (?, ?, ?, ?, ?, ?, NULL)
	`, "arc0000000000000000000001", "org1", "acct1", now, now, "appbranchrun000000000001").Error)

	a := &Activities{
		v:  validator.New(),
		db: db,
		l:  zap.NewNop(),
	}

	out, err := a.ComputeAndStoreAppBranchRunComparison(context.Background(), &ComputeAndStoreAppBranchRunComparisonInput{
		AppBranchID: "appbranch0000000000000001",
		RunID:       "appbranchrun000000000001",
	})
	require.NoError(t, err)
	require.True(t, out.Skipped)
	require.Equal(t, "no base run", out.SkipReason)
}

func TestRunCommitSHA(t *testing.T) {
	sha := "abc123"
	require.Equal(t, sha, runCommitSHA(&app.AppBranchRun{
		VCSConnectionCommit: &app.VCSConnectionCommit{SHA: sha},
		HeadSHA:             "ignored",
	}))
	require.Equal(t, "head", runCommitSHA(&app.AppBranchRun{HeadSHA: "head"}))
	require.Equal(t, "", runCommitSHA(&app.AppBranchRun{}))
}
