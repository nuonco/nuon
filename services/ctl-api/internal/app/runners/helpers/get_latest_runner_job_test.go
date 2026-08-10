package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestGetLatestJobCompositeError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE runner_jobs (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			owner_type TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			deleted_at INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE runner_job_executions (
			id TEXT PRIMARY KEY,
			runner_job_id TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			deleted_at INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE runner_job_execution_results (
			id TEXT PRIMARY KEY,
			runner_job_execution_id TEXT NOT NULL,
			composite_error TEXT,
			deleted_at INTEGER NOT NULL DEFAULT 0
		);
	`).Error)

	now := time.Now()
	errorJSON := `{"version":1,"type":"aws_permission_error","severity":"error","message":"missing permission","data":{}}`
	require.NoError(t, db.Exec(`
		INSERT INTO runner_jobs (id, owner_id, owner_type, created_at) VALUES
			('old-job', 'owner-with-new-success', 'component_builds', ?),
			('new-job', 'owner-with-new-success', 'component_builds', ?),
			('error-job', 'owner-with-error', 'component_builds', ?),
			('pending-job', 'owner-with-pending-retry', 'component_builds', ?),
			('execution-retry-job', 'owner-with-execution-retry', 'component_builds', ?),
			('wrong-owner-type-job', 'owner-with-error', 'install_action_workflow_runs', ?);
		INSERT INTO runner_job_executions (id, runner_job_id, created_at) VALUES
			('old-failed-execution', 'old-job', ?),
			('new-success-execution', 'new-job', ?),
			('older-error-execution', 'error-job', ?),
			('newer-error-execution', 'error-job', ?),
			('pending-execution', 'pending-job', ?),
			('retry-failed-execution', 'execution-retry-job', ?),
			('retry-success-execution', 'execution-retry-job', ?),
			('wrong-owner-type-execution', 'wrong-owner-type-job', ?);
		INSERT INTO runner_job_execution_results (id, runner_job_execution_id, composite_error) VALUES
			('old-failed-result', 'old-failed-execution', ?),
			('new-success-result', 'new-success-execution', NULL),
			('older-error-result', 'older-error-execution', NULL),
			('newer-error-result', 'newer-error-execution', ?),
			('retry-failed-result', 'retry-failed-execution', ?),
			('retry-success-result', 'retry-success-execution', NULL),
			('wrong-owner-type-result', 'wrong-owner-type-execution', NULL);
	`,
		now.Add(-2*time.Hour), now.Add(-time.Hour), now, now, now, now.Add(time.Hour),
		now.Add(-2*time.Hour), now.Add(-time.Hour), now.Add(-time.Hour), now, now,
		now.Add(-time.Hour), now, now.Add(time.Hour),
		errorJSON, errorJSON, errorJSON,
	).Error)

	ctx := context.Background()
	compositeError, err := GetLatestJobCompositeError(ctx, db, GetLatestJobCompositeErrorRequest{
		OwnerID:   "owner-with-error",
		OwnerType: "component_builds",
	})
	require.NoError(t, err)
	require.Equal(t, compositeerrors.Type("aws_permission_error"), compositeError.Type)

	compositeError, err = GetLatestJobCompositeError(ctx, db, GetLatestJobCompositeErrorRequest{
		OwnerID:   "owner-with-new-success",
		OwnerType: "component_builds",
	})
	require.NoError(t, err)
	require.Nil(t, compositeError)

	compositeError, err = GetLatestJobCompositeError(ctx, db, GetLatestJobCompositeErrorRequest{
		OwnerID:   "owner-with-pending-retry",
		OwnerType: "component_builds",
	})
	require.NoError(t, err)
	require.Nil(t, compositeError)

	compositeError, err = GetLatestJobCompositeError(ctx, db, GetLatestJobCompositeErrorRequest{
		OwnerID:   "owner-with-execution-retry",
		OwnerType: "component_builds",
	})
	require.NoError(t, err)
	require.Nil(t, compositeError)

	compositeError, err = GetLatestJobCompositeError(ctx, db, GetLatestJobCompositeErrorRequest{
		OwnerID:   "missing-owner",
		OwnerType: "component_builds",
	})
	require.NoError(t, err)
	require.Nil(t, compositeError)
}
