package processjob

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/joberrors"
	runneractivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func executeProcessJobSignal(t *testing.T, env *testsuite.TestWorkflowEnvironment) error {
	t.Helper()
	sig := &Signal{RunnerID: "runner-1", JobID: "job-1"}
	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })
	require.True(t, env.IsWorkflowCompleted())
	return env.GetWorkflowError()
}

func newProcessJobTestEnvironment() *testsuite.TestWorkflowEnvironment {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})
	return env
}

func TestExecuteRecordsNoActiveRunnerCompositeError(t *testing.T) {
	env := newProcessJobTestEnvironment()
	env.RegisterActivityWithOptions((&runneractivities.Activities{}).RecordJobLifecycleCompositeError,
		activity.RegisterOptions{Name: "RecordJobLifecycleCompositeError"})
	env.OnActivity((*runneractivities.Activities).GetRunnerJobForExecution, mock.Anything, mock.Anything, mock.Anything).
		Return(&runneractivities.GetRunnerJobForExecutionResponse{
			Job: &app.RunnerJob{ID: "job-1", Status: app.RunnerJobStatusQueued},
		}, nil).Once()
	env.OnActivity((*runneractivities.Activities).UpdateJobStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()
	env.OnActivity((*statusactivities.Activities).UpdateRunnerJobStatusV2, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()
	env.OnActivity("RecordJobLifecycleCompositeError", mock.Anything,
		mock.MatchedBy(func(req runneractivities.RecordJobLifecycleCompositeErrorRequest) bool {
			return req.JobID == "job-1" && req.Reason == joberrors.LifecycleFailureReasonNoActiveRunner
		})).Return(errors.New("composite error write failed")).Once()

	require.ErrorContains(t, executeProcessJobSignal(t, env), "runner has no active process")
	env.AssertExpectations(t)
}

func TestExecuteDoesNotRecordLifecycleErrorForCancelledJob(t *testing.T) {
	env := newProcessJobTestEnvironment()
	env.OnActivity((*runneractivities.Activities).GetRunnerJobForExecution, mock.Anything, mock.Anything, mock.Anything).
		Return(&runneractivities.GetRunnerJobForExecutionResponse{
			Job: &app.RunnerJob{ID: "job-1", Status: app.RunnerJobStatusCancelled},
		}, nil).Once()

	require.NoError(t, executeProcessJobSignal(t, env))
	env.AssertExpectations(t)
	env.AssertNumberOfCalls(t, "UpdateJobStatus", 0)
	env.AssertNumberOfCalls(t, "RecordJobLifecycleCompositeError", 0)
}
