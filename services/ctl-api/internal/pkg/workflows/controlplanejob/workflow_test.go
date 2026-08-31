package controlplanejob

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestExecuteControlPlaneJobDoesNotRetryJobFailure(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	env.OnActivity((*Activities).ControlPlaneJobEnsureExecution, mock.Anything, mock.Anything, mock.Anything).
		Return(&EnsureExecutionResponse{ExecutionID: "execution-id", JobExecutionTimeout: time.Minute}, nil).
		Once()
	env.OnActivity((*Activities).ControlPlaneJobRunJob, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("image not found")).
		Once()
	env.OnActivity((*Activities).ControlPlaneJobFinalize, mock.Anything, mock.Anything, mock.Anything).
		Return(&FinalizeResponse{Status: app.RunnerJobExecutionStatusFailed}, nil).
		Once()

	env.ExecuteWorkflow((&Workflows{}).ExecuteControlPlaneJob, &ExecuteRequest{JobID: "job-id"})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}
