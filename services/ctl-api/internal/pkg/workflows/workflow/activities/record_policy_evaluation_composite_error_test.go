package activities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/policyerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestRecordPolicyEvaluationCompositeErrorForRunnerJob(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE runner_jobs (
			id TEXT PRIMARY KEY,
			updated_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			composite_error TEXT
		);
		INSERT INTO runner_jobs (id) VALUES ('job-1');
	`).Error)

	activities := &Activities{db: db}
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(activities.RecordPolicyEvaluationCompositeError)

	_, err = env.ExecuteActivity(activities.RecordPolicyEvaluationCompositeError, RecordPolicyEvaluationCompositeErrorRequest{
		RunnerJobID: "job-1",
		Stage:       policyerrors.EvaluationFailureStageEvaluation,
	})
	require.NoError(t, err)

	var raw string
	require.NoError(t, db.Raw(`SELECT composite_error FROM runner_jobs WHERE id = 'job-1'`).Scan(&raw).Error)
	var data compositeerrors.CompositeErrorData
	require.NoError(t, json.Unmarshal([]byte(raw), &data))
	require.Equal(t, policyerrors.EvaluationFailedErrorType, data.Type)
	require.Equal(t, compositeerrors.SeverityWarning, data.Severity)
	require.Equal(t, "runner_jobs", data.SourceType)
	require.Equal(t, "job-1", data.SourceID)
}
