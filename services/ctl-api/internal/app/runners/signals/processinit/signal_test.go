package processinit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func TestProcessInitPreservesStackStartupTransition(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

	sig := &Signal{RunnerID: "rnr_1", ProcessID: "rnp_1"}
	process := &app.RunnerProcess{
		ID:   sig.ProcessID,
		Type: app.RunnerProcessTypeInstall,
		CompositeStatus: app.CompositeStatus{
			Status: app.StatusPending,
		},
	}
	runner := &app.Runner{
		ID:     sig.RunnerID,
		Status: app.RunnerStatusAwaitingInstallStackRun,
		StatusV2: app.CompositeStatus{
			Status: app.Status(app.RunnerStatusAwaitingInstallStackRun),
		},
	}

	env.OnActivity((*activities.Activities).GetRunnerProcess, mock.Anything, mock.Anything, mock.Anything).
		Return(process, nil).
		Once()
	env.OnActivity((*activities.Activities).UpdateRunnerProcessStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(process, nil).
		Once()
	env.OnActivity((*activities.Activities).Get, mock.Anything, mock.Anything, mock.Anything).
		Return(runner, nil).
		Once()
	var transitions []activities.UpdateStatusRequest
	env.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, args converter.EncodedValues) {
		if info.ActivityType.Name != "UpdateStatus" {
			return
		}
		var req activities.UpdateStatusRequest
		require.NoError(t, args.Get(&req))
		transitions = append(transitions, req)
	})
	env.OnActivity((*activities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Twice()
	env.OnActivity((*activities.Activities).GetStaleRunnerProcesses, mock.Anything, mock.Anything, mock.Anything).
		Return([]app.RunnerProcess{}, nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Execute(ctx)
	})

	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, []activities.UpdateStatusRequest{
		{
			RunnerID:          sig.RunnerID,
			Status:            app.RunnerStatusAwaitingHeartbeat,
			StatusDescription: "runner install stack was run, waiting for the runner to report in",
		},
		{
			RunnerID:          sig.RunnerID,
			Status:            app.RunnerStatusActive,
			StatusDescription: "process initialized",
			Metadata: map[string]any{
				app.RunnerOfflineTSMetadataKey: nil,
			},
		},
	}, transitions)
	env.AssertExpectations(t)
}

func TestExistingHistoryKeepsSplitStatusWrites(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})
	env.OnGetVersion(atomicRunnerStatusVersion, workflow.DefaultVersion, 1).
		Return(workflow.DefaultVersion).
		Once()

	sig := &Signal{RunnerID: "rnr_1", ProcessID: "rnp_1"}
	process := &app.RunnerProcess{
		ID:   sig.ProcessID,
		Type: app.RunnerProcessTypeInstall,
		CompositeStatus: app.CompositeStatus{
			Status: app.StatusPending,
		},
	}
	env.OnActivity((*activities.Activities).GetRunnerProcess, mock.Anything, mock.Anything, mock.Anything).
		Return(process, nil).
		Once()
	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2Metadata, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()
	env.OnActivity((*activities.Activities).UpdateRunnerProcessStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(process, nil).
		Once()
	env.OnActivity((*activities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()
	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()
	env.OnActivity((*activities.Activities).GetStaleRunnerProcesses, mock.Anything, mock.Anything, mock.Anything).
		Return([]app.RunnerProcess{}, nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Execute(ctx)
	})

	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}
