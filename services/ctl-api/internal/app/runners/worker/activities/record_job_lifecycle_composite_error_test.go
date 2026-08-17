package activities

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/joberrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestRecordJobLifecycleCompositeError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE runner_jobs (
			id TEXT PRIMARY KEY,
			owner_id TEXT,
			owner_type TEXT,
			created_at DATETIME NOT NULL,
			status TEXT NOT NULL,
			updated_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			composite_error TEXT
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
		CREATE TABLE install_deploys (
			id TEXT PRIMARY KEY,
			updated_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			composite_error TEXT
		);
	`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO install_deploys (id) VALUES ('deploy-1'), ('deploy-with-result');
		INSERT INTO runner_jobs (id, owner_id, owner_type, created_at, status) VALUES
			('job-1', 'deploy-1', 'install_deploys', CURRENT_TIMESTAMP, 'timed-out'),
			('job-with-owner-result', 'deploy-with-result', 'install_deploys', CURRENT_TIMESTAMP, 'failed'),
			('job-with-result', NULL, NULL, CURRENT_TIMESTAMP, 'failed'),
			('job-without-result', NULL, NULL, CURRENT_TIMESTAMP, 'failed');
		INSERT INTO runner_job_executions (id, runner_job_id, created_at) VALUES
			('execution-with-owner-result', 'job-with-owner-result', CURRENT_TIMESTAMP),
			('execution-with-result', 'job-with-result', CURRENT_TIMESTAMP),
			('execution-without-result', 'job-without-result', CURRENT_TIMESTAMP);
		INSERT INTO runner_job_execution_results (id, runner_job_execution_id, composite_error) VALUES
			('owner-result', 'execution-with-owner-result', ?),
			('result-1', 'execution-with-result', NULL);
	`, `{"version":1,"type":"terraform.aws_permission","severity":"error","message":"missing permission","data":{}}`).Error)

	activities := &Activities{db: db}
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(activities.RecordJobLifecycleCompositeError)

	_, err = env.ExecuteActivity(activities.RecordJobLifecycleCompositeError, RecordJobLifecycleCompositeErrorRequest{
		JobID:  "job-1",
		Reason: joberrors.LifecycleFailureReasonPickupTimeout,
	})
	require.NoError(t, err)

	var raw sql.NullString
	require.NoError(t, db.Raw(`SELECT composite_error FROM runner_jobs WHERE id = 'job-1'`).Scan(&raw).Error)
	require.True(t, raw.Valid)

	var data compositeerrors.CompositeErrorData
	require.NoError(t, json.Unmarshal([]byte(raw.String), &data))
	require.Equal(t, joberrors.LifecycleFailureErrorType, data.Type)
	require.Equal(t, "runner_jobs", data.SourceType)
	require.Equal(t, "job-1", data.SourceID)

	require.NoError(t, db.Raw(`SELECT composite_error FROM install_deploys WHERE id = 'deploy-1'`).Scan(&raw).Error)
	require.True(t, raw.Valid)
	require.NoError(t, json.Unmarshal([]byte(raw.String), &data))
	require.Equal(t, joberrors.LifecycleFailureErrorType, data.Type)

	_, err = env.ExecuteActivity(activities.RecordJobLifecycleCompositeError, RecordJobLifecycleCompositeErrorRequest{
		JobID:  "job-with-owner-result",
		Reason: joberrors.LifecycleFailureReasonRunnerUnhealthy,
	})
	require.NoError(t, err)
	require.NoError(t, db.Raw(`SELECT composite_error FROM install_deploys WHERE id = 'deploy-with-result'`).Scan(&raw).Error)
	require.True(t, raw.Valid)
	require.NoError(t, json.Unmarshal([]byte(raw.String), &data))
	require.Equal(t, compositeerrors.Type("terraform.aws_permission"), data.Type)

	_, err = env.ExecuteActivity(activities.RecordJobLifecycleCompositeError, RecordJobLifecycleCompositeErrorRequest{
		JobID:  "job-with-result",
		Reason: joberrors.LifecycleFailureReasonResultMissing,
	})
	require.NoError(t, err)
	require.NoError(t, db.Raw(`SELECT composite_error FROM runner_jobs WHERE id = 'job-with-result'`).Scan(&raw).Error)
	require.False(t, raw.Valid)

	_, err = env.ExecuteActivity(activities.RecordJobLifecycleCompositeError, RecordJobLifecycleCompositeErrorRequest{
		JobID:  "job-without-result",
		Reason: joberrors.LifecycleFailureReasonResultMissing,
	})
	require.NoError(t, err)
	require.NoError(t, db.Raw(`SELECT composite_error FROM runner_jobs WHERE id = 'job-without-result'`).Scan(&raw).Error)
	require.True(t, raw.Valid)
	require.NoError(t, json.Unmarshal([]byte(raw.String), &data))
	var lifecycleError joberrors.LifecycleFailureError
	require.NoError(t, json.Unmarshal(data.Data, &lifecycleError))
	require.Equal(t, joberrors.LifecycleFailureReasonResultMissing, lifecycleError.Reason)

	_, err = env.ExecuteActivity(activities.RecordJobLifecycleCompositeError, RecordJobLifecycleCompositeErrorRequest{
		JobID:  "missing-job",
		Reason: joberrors.LifecycleFailureReasonQueueTimeout,
	})
	require.Error(t, err)
}
