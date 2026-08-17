package activities

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/actions/actionerrors"
)

func TestSetActionWorkflowRunPreparationCompositeError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE install_action_workflow_runs (
			id TEXT PRIMARY KEY,
			updated_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			composite_error TEXT
		);
		INSERT INTO install_action_workflow_runs (id) VALUES ('run-1');
	`).Error)

	activities := &Activities{db: db}
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(activities.SetActionWorkflowRunPreparationCompositeError)

	_, err = env.ExecuteActivity(activities.SetActionWorkflowRunPreparationCompositeError, SetActionWorkflowRunPreparationCompositeErrorRequest{
		RunID:  "run-1",
		Detail: "api_token=secret-value could not be resolved",
	})
	require.NoError(t, err)

	var run app.InstallActionWorkflowRun
	require.NoError(t, db.Where(app.InstallActionWorkflowRun{ID: "run-1"}).First(&run).Error)
	require.NotNil(t, run.CompositeError)
	require.Equal(t, actionerrors.PreparationFailedErrorType, run.CompositeError.Type)
	require.Equal(t, "install_action_workflow_runs", run.CompositeError.SourceType)
	require.Equal(t, "run-1", run.CompositeError.SourceID)
	require.Contains(t, run.CompositeError.Sections[1].Body, "api_token=[REDACTED]")
	require.NotContains(t, run.CompositeError.Sections[1].Body, "secret-value")

	_, err = env.ExecuteActivity(activities.SetActionWorkflowRunPreparationCompositeError, SetActionWorkflowRunPreparationCompositeErrorRequest{
		RunID: "run-1",
	})
	require.NoError(t, err)

	run = app.InstallActionWorkflowRun{}
	require.NoError(t, db.Where(app.InstallActionWorkflowRun{ID: "run-1"}).First(&run).Error)
	require.Nil(t, run.CompositeError)

	_, err = env.ExecuteActivity(activities.SetActionWorkflowRunPreparationCompositeError, SetActionWorkflowRunPreparationCompositeErrorRequest{
		RunID:  "missing-run",
		Detail: "preparation failed",
	})
	require.ErrorContains(t, err, gorm.ErrRecordNotFound.Error())
}
