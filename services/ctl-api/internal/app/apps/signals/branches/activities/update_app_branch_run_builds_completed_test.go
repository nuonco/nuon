package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func setupBuildsCompletedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE app_branch_runs (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL DEFAULT '',
			app_branch_id TEXT NOT NULL DEFAULT '',
			app_branch_config_id TEXT NOT NULL DEFAULT '',
			created_by_id TEXT NOT NULL DEFAULT '',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			run_type TEXT NOT NULL DEFAULT 'manual-run',
			plan_only INTEGER NOT NULL DEFAULT 0,
			force INTEGER NOT NULL DEFAULT 0,
			labels TEXT,
			no_config_changes INTEGER NOT NULL DEFAULT 0,
			error_message TEXT NOT NULL DEFAULT '',
			app_config_id TEXT NOT NULL DEFAULT '',
			head_sha TEXT NOT NULL DEFAULT '',
			base_branch TEXT NOT NULL DEFAULT '',
			event_type TEXT NOT NULL DEFAULT ''
		)
	`).Error)
	return db
}

func TestUpdateAppBranchRunBuildsCompleted(t *testing.T) {
	db := setupBuildsCompletedDB(t)
	require.NoError(t, db.Exec(`
		INSERT INTO app_branch_runs (id, labels, created_at, updated_at)
		VALUES ('run-1', '{"commit":"abc"}', datetime('now'), datetime('now'))
	`).Error)

	a := &Activities{db: db, v: validator.New()}
	require.NoError(t, a.UpdateAppBranchRunBuildsCompleted(context.Background(), &UpdateAppBranchRunBuildsCompletedInput{
		RunID:           "run-1",
		BuildsCompleted: true,
	}))

	var run app.AppBranchRun
	require.NoError(t, db.First(&run, "id = ?", "run-1").Error)
	require.Equal(t, "true", run.Labels[app.AppBranchRunLabelBuildsCompleted])
	require.Equal(t, "abc", run.Labels["commit"])

	require.NoError(t, a.UpdateAppBranchRunBuildsCompleted(context.Background(), &UpdateAppBranchRunBuildsCompletedInput{
		RunID:           "run-1",
		BuildsCompleted: false,
	}))
	require.NoError(t, db.First(&run, "id = ?", "run-1").Error)
	require.Equal(t, "false", run.Labels[app.AppBranchRunLabelBuildsCompleted])
	require.Equal(t, "abc", run.Labels["commit"])
}

func TestUpdateAppBranchRunBuildsCompletedNilLabels(t *testing.T) {
	db := setupBuildsCompletedDB(t)
	require.NoError(t, db.Exec(`
		INSERT INTO app_branch_runs (id, labels, created_at, updated_at)
		VALUES ('run-2', NULL, datetime('now'), datetime('now'))
	`).Error)

	a := &Activities{db: db, v: validator.New()}
	require.NoError(t, a.UpdateAppBranchRunBuildsCompleted(context.Background(), &UpdateAppBranchRunBuildsCompletedInput{
		RunID:           "run-2",
		BuildsCompleted: true,
	}))

	var run app.AppBranchRun
	require.NoError(t, db.First(&run, "id = ?", "run-2").Error)
	require.True(t, run.Labels.HasLabel(app.AppBranchRunLabelBuildsCompleted, "true"))
}
