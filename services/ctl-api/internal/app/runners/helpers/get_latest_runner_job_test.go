package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/joberrors"
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
			status TEXT NOT NULL,
			composite_error TEXT,
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
	lifecycleErrorJSON := `{"version":1,"type":"runner.job_lifecycle_failure","severity":"error","message":"runner unavailable","data":{"reason":"runner_unhealthy"}}`
	cancellationErrorJSON := `{"version":1,"type":"runner.job_cancelled","severity":"warning","message":"job cancelled","data":{"reason":"api"}}`
	require.NoError(t, db.Exec(`
		INSERT INTO runner_jobs (id, owner_id, owner_type, created_at, status, composite_error) VALUES
			('old-job', 'owner-with-new-success', 'component_builds', ?, 'failed', ?),
			('new-job', 'owner-with-new-success', 'component_builds', ?, 'finished', NULL),
			('error-job', 'owner-with-error', 'component_builds', ?, 'failed', ?),
			('pending-job', 'owner-with-pending-retry', 'component_builds', ?, 'available', ?),
			('execution-retry-job', 'owner-with-execution-retry', 'component_builds', ?, 'failed', NULL),
			('lifecycle-failed-job', 'owner-with-lifecycle-failure', 'component_builds', ?, 'failed', ?),
			('lifecycle-timed-out-job', 'owner-with-lifecycle-timeout', 'component_builds', ?, 'timed-out', ?),
			('lifecycle-not-attempted-job', 'owner-with-lifecycle-not-attempted', 'component_builds', ?, 'not-attempted', ?),
			('finished-lifecycle-job', 'owner-with-finished-lifecycle', 'component_builds', ?, 'finished', ?),
			('cancelled-lifecycle-job', 'owner-with-cancelled-lifecycle', 'component_builds', ?, 'cancelled', ?),
			('cancelled-api-job', 'owner-with-api-cancellation', 'component_builds', ?, 'cancelled', ?),
			('old-lifecycle-retry-job', 'owner-with-new-pending-job', 'component_builds', ?, 'failed', ?),
			('new-pending-job', 'owner-with-new-pending-job', 'component_builds', ?, 'queued', NULL),
			('wrong-owner-type-job', 'owner-with-error', 'install_action_workflow_runs', ?, 'finished', NULL);
	`,
		now.Add(-2*time.Hour), lifecycleErrorJSON,
		now.Add(-time.Hour),
		now, lifecycleErrorJSON,
		now, lifecycleErrorJSON,
		now,
		now, lifecycleErrorJSON,
		now, lifecycleErrorJSON,
		now, lifecycleErrorJSON,
		now, lifecycleErrorJSON,
		now, lifecycleErrorJSON,
		now, cancellationErrorJSON,
		now.Add(-time.Hour), lifecycleErrorJSON,
		now,
		now.Add(time.Hour),
	).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO runner_job_executions (id, runner_job_id, created_at) VALUES
			('old-failed-execution', 'old-job', ?),
			('new-success-execution', 'new-job', ?),
			('older-error-execution', 'error-job', ?),
			('newer-error-execution', 'error-job', ?),
			('pending-execution', 'pending-job', ?),
			('retry-failed-execution', 'execution-retry-job', ?),
			('retry-success-execution', 'execution-retry-job', ?),
			('lifecycle-timeout-execution', 'lifecycle-timed-out-job', ?),
			('finished-stale-execution', 'finished-lifecycle-job', ?),
			('wrong-owner-type-execution', 'wrong-owner-type-job', ?);
	`,
		now.Add(-2*time.Hour), now.Add(-time.Hour), now.Add(-time.Hour), now, now,
		now.Add(-time.Hour), now, now, now, now.Add(time.Hour),
	).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO runner_job_execution_results (id, runner_job_execution_id, composite_error) VALUES
			('old-failed-result', 'old-failed-execution', ?),
			('new-success-result', 'new-success-execution', NULL),
			('older-error-result', 'older-error-execution', NULL),
			('newer-error-result', 'newer-error-execution', ?),
			('pending-error-result', 'pending-execution', ?),
			('retry-failed-result', 'retry-failed-execution', ?),
			('retry-success-result', 'retry-success-execution', NULL),
			('lifecycle-timeout-result', 'lifecycle-timeout-execution', NULL),
			('finished-stale-result', 'finished-stale-execution', ?),
			('wrong-owner-type-result', 'wrong-owner-type-execution', NULL);
	`, errorJSON, errorJSON, errorJSON, errorJSON, errorJSON,
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

	for _, ownerID := range []string{
		"owner-with-lifecycle-failure",
		"owner-with-lifecycle-timeout",
		"owner-with-lifecycle-not-attempted",
	} {
		compositeError, err = GetLatestJobCompositeError(ctx, db, GetLatestJobCompositeErrorRequest{
			OwnerID:   ownerID,
			OwnerType: "component_builds",
		})
		require.NoError(t, err)
		require.NotNil(t, compositeError)
		require.Equal(t, compositeerrors.Type("runner.job_lifecycle_failure"), compositeError.Type)
	}

	compositeError, err = GetLatestJobCompositeError(ctx, db, GetLatestJobCompositeErrorRequest{
		OwnerID:   "owner-with-api-cancellation",
		OwnerType: "component_builds",
	})
	require.NoError(t, err)
	require.NotNil(t, compositeError)
	require.Equal(t, joberrors.CancellationErrorType, compositeError.Type)

	for _, ownerID := range []string{
		"owner-with-finished-lifecycle",
		"owner-with-cancelled-lifecycle",
		"owner-with-new-pending-job",
	} {
		compositeError, err = GetLatestJobCompositeError(ctx, db, GetLatestJobCompositeErrorRequest{
			OwnerID:   ownerID,
			OwnerType: "component_builds",
		})
		require.NoError(t, err)
		require.Nil(t, compositeError)
	}

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

func TestResolveJobCompositeError(t *testing.T) {
	executionError := &compositeerrors.CompositeErrorData{Type: "terraform.error"}
	lifecycleError := &compositeerrors.CompositeErrorData{Type: "runner.job_lifecycle_failure"}
	cancellationError := &compositeerrors.CompositeErrorData{Type: joberrors.CancellationErrorType}

	tests := map[string]struct {
		job      app.RunnerJob
		expected *compositeerrors.CompositeErrorData
	}{
		"latest execution result takes precedence": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusFailed,
				CompositeError: lifecycleError,
				Executions: []app.RunnerJobExecution{{
					Result: &app.RunnerJobExecutionResult{CompositeError: executionError},
				}},
			},
			expected: executionError,
		},
		"failed job uses lifecycle error": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusFailed,
				CompositeError: lifecycleError,
			},
			expected: lifecycleError,
		},
		"cancelled job uses cancellation error instead of execution error": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusCancelled,
				CompositeError: cancellationError,
				Executions: []app.RunnerJobExecution{{
					Result: &app.RunnerJobExecutionResult{CompositeError: executionError},
				}},
			},
			expected: cancellationError,
		},
		"cancelled job hides stale lifecycle error": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusCancelled,
				CompositeError: lifecycleError,
			},
		},
		"not attempted job uses lifecycle error": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusNotAttempted,
				CompositeError: lifecycleError,
			},
			expected: lifecycleError,
		},
		"finished job hides stale lifecycle error": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusFinished,
				CompositeError: lifecycleError,
			},
		},
		"finished job hides stale execution error": {
			job: app.RunnerJob{
				Status: app.RunnerJobStatusFinished,
				Executions: []app.RunnerJobExecution{{
					Result: &app.RunnerJobExecutionResult{CompositeError: executionError},
				}},
			},
		},
		"retrying job hides previous execution error": {
			job: app.RunnerJob{
				Status: app.RunnerJobStatusAvailable,
				Executions: []app.RunnerJobExecution{{
					Result: &app.RunnerJobExecutionResult{CompositeError: executionError},
				}},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Same(t, tt.expected, ResolveJobCompositeError(&tt.job))
		})
	}
}
