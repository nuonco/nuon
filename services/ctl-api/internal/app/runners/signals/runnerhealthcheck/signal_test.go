package runnerhealthcheck

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runneractivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func TestUpdateRunnerStatusEnqueuesUnhealthyAfterOfflineTransition(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	runner := testRunner(app.RunnerStatusActive)
	sig := &Signal{RunnerID: runner.ID}
	var calls []string

	env.OnActivity((*runneractivities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { calls = append(calls, "status") }).
		Return(nil).
		Once()
	env.OnActivity((*sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) {
			calls = append(calls, "notification")
		}).
		Return(&sharedactivities.EnqueueSignalToOwnerResponse{}, nil).
		Once()
	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { calls = append(calls, "status-v2") }).
		Return(nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.updateRunnerStatus(ctx, runner, app.RunnerStatusOffline, "no active build process", runner.Org.Name, nil)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, []string{"status", "notification", "status-v2"}, calls)
	env.AssertExpectations(t)
}

func TestUpdateRunnerStatusDoesNotRepeatUnhealthyWhileOffline(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	runner := testRunner(app.RunnerStatusOffline)
	sig := &Signal{RunnerID: runner.ID}

	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.updateRunnerStatus(ctx, runner, app.RunnerStatusOffline, "no active build process", runner.Org.Name, nil)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestUpdateRunnerStatusDoesNotEnqueueWhenOfflineUpdateFails(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	runner := testRunner(app.RunnerStatusActive)
	sig := &Signal{RunnerID: runner.ID}

	env.OnActivity((*runneractivities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(temporal.NewNonRetryableApplicationError("update failed", "test", nil)).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.updateRunnerStatus(ctx, runner, app.RunnerStatusOffline, "no active build process", runner.Org.Name, nil)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func testRunner(status app.RunnerStatus) *app.Runner {
	return &app.Runner{
		ID:          "rnr_1",
		DisplayName: "Install runner",
		OrgID:       "org_1",
		Org: app.Org{
			Name: "Example org",
		},
		Status:        status,
		RunnerGroupID: "rng_1",
		RunnerGroup: app.RunnerGroup{
			Type:      app.RunnerGroupTypeInstall,
			OwnerID:   "ins_1",
			OwnerType: "installs",
		},
	}
}
