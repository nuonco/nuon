package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestUpdateJobStatusOnlyUpdatesQueuedJob(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE runner_jobs (
			id TEXT PRIMARY KEY,
			status TEXT,
			status_description TEXT,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0
		)
	`).Error)

	insertJob := func(t *testing.T, id string, status app.RunnerJobStatus) {
		t.Helper()
		require.NoError(t, db.Exec(
			`INSERT INTO runner_jobs (id, status, status_description) VALUES (?, ?, '')`,
			id,
			status,
		).Error)
	}
	insertJob(t, "queued-job", app.RunnerJobStatusQueued)
	insertJob(t, "in-progress-job", app.RunnerJobStatusInProgress)

	a := &Activities{db: db}
	require.NoError(t, a.UpdateJobStatus(context.Background(), &UpdateJobStatusRequest{
		JobID:             "queued-job",
		Status:            app.RunnerJobStatusFailed,
		StatusDescription: "build setup failed",
		ExpectedStatus:    app.RunnerJobStatusQueued,
	}))

	require.NoError(t, a.UpdateJobStatus(context.Background(), &UpdateJobStatusRequest{
		JobID:             "in-progress-job",
		Status:            app.RunnerJobStatusFailed,
		StatusDescription: "stale setup failure",
		ExpectedStatus:    app.RunnerJobStatusQueued,
	}))

	var queuedJob app.RunnerJob
	require.NoError(t, db.Where(app.RunnerJob{ID: "queued-job"}).First(&queuedJob).Error)
	require.Equal(t, app.RunnerJobStatusFailed, queuedJob.Status)
	require.Equal(t, "build setup failed", queuedJob.StatusDescription)

	var inProgressJob app.RunnerJob
	require.NoError(t, db.Where(app.RunnerJob{ID: "in-progress-job"}).First(&inProgressJob).Error)
	require.Equal(t, app.RunnerJobStatusInProgress, inProgressJob.Status)
	require.Empty(t, inProgressJob.StatusDescription)
}
