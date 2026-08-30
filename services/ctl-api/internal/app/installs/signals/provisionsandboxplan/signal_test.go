package provisionsandboxplan

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type SignalTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestSignalTestSuite(t *testing.T) {
	suite.Run(t, new(SignalTestSuite))
}

func (s *SignalTestSuite) TestExecute() {
	s.T().Skip("not yet implemented")
}

// TestExecutePlanFailureRecordsCompositeError verifies that when the sandbox
// plan child workflow fails, a SandboxPlanRenderError is written to the
// sandbox run row and a non-retryable error is returned to the caller.
func (s *SignalTestSuite) TestExecutePlanFailureRecordsCompositeError() {
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

	sig := &Signal{
		InstallSandboxID: "sandbox-1",
		cfg:              &internal.Config{},
	}

	install := &app.Install{
		ID:       "install-1",
		RunnerID: "runner-1",
		AppSandboxConfig: app.AppSandboxConfig{
			ID: "sandbox-config-1",
		},
	}
	sandboxRun := &app.InstallSandboxRun{ID: "run-1", RunType: app.SandboxRunTypeProvision}
	logStream := &app.LogStream{ID: "log-1"}
	runnerJob := &app.RunnerJob{ID: "job-1"}

	env.OnActivity((*activities.Activities).GetInstallForSandbox, mock.Anything, mock.Anything, mock.Anything).
		Return(install, nil)
	env.OnActivity((*activities.Activities).CreateSandboxRun, mock.Anything, mock.Anything, mock.Anything).
		Return(sandboxRun, nil)
	env.OnActivity((*activities.Activities).SetInstallSandboxRunPlanCompositeError, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	env.OnActivity((*activities.Activities).UpdateRunStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	env.OnActivity((*statusactivities.Activities).UpdateRunStatusV2, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	env.OnActivity((*activities.Activities).CreateLogStream, mock.Anything, mock.Anything, mock.Anything).
		Return(logStream, nil)
	env.OnActivity((*activities.Activities).CreateSandboxJob, mock.Anything, mock.Anything, mock.Anything).
		Return(runnerJob, nil)

	planErr := temporal.NewNonRetryableApplicationError("env var {{.Missing}} undefined", "RENDER_FAILED", nil)
	env.OnWorkflow(plan.CreateSandboxRunPlan, mock.Anything, mock.Anything).
		Return(plan.CreateSandboxPlanResponse{}, planErr)
	env.OnActivity((*activities.Activities).CloseLogStream, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	var compositeErrCalls []activities.SetInstallSandboxRunPlanCompositeErrorRequest
	env.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, args converter.EncodedValues) {
		if info.ActivityType.Name == "SetInstallSandboxRunPlanCompositeError" {
			var req activities.SetInstallSandboxRunPlanCompositeErrorRequest
			if err := args.Get(&req); err == nil {
				compositeErrCalls = append(compositeErrCalls, req)
			}
		}
	})

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Execute(ctx)
	})

	require.Error(s.T(), env.GetWorkflowError())
	var appErr *temporal.ApplicationError
	require.ErrorAs(s.T(), env.GetWorkflowError(), &appErr)
	require.True(s.T(), appErr.NonRetryable(), "plan render failure must produce a non-retryable error")

	require.GreaterOrEqual(s.T(), len(compositeErrCalls), 2, "expected clear-on-start and error-recording calls")
	var errorCall *activities.SetInstallSandboxRunPlanCompositeErrorRequest
	for i := range compositeErrCalls {
		if compositeErrCalls[i].Detail != "" {
			errorCall = &compositeErrCalls[i]
		}
	}
	require.NotNil(s.T(), errorCall, "expected a SetInstallSandboxRunPlanCompositeError call with non-empty Detail")
	require.Equal(s.T(), "run-1", errorCall.SandboxRunID)
}
