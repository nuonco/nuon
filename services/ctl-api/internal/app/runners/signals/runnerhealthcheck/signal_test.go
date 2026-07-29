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

func TestSecondFailedHealthCheckEnqueuesUnhealthyAfterOfflineTransition(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	runner := runnerWithHealthCheckFailures(app.RunnerStatusActive, float64(1))
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
		status, metadata := runnerStatusAfterHealthCheck(runner, app.RunnerStatusOffline, nil)
		return sig.updateRunnerStatus(ctx, runner, status, "no active build process", runner.Org.Name, metadata)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, []string{"status", "notification", "status-v2"}, calls)
	env.AssertExpectations(t)
}

func TestFirstFailedHealthCheckDoesNotMarkRunnerOffline(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	runner := testRunner(app.RunnerStatusActive)
	sig := &Signal{RunnerID: runner.ID}

	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2, mock.MatchedBy(func(req statusactivities.UpdateRunnerStatusV2Request) bool {
		return req.Status == app.RunnerStatusActive && req.Metadata[app.RunnerHealthCheckConsecutiveFailuresMetadataKey] == 1
	})).Return(nil).Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		status, metadata := runnerStatusAfterHealthCheck(runner, app.RunnerStatusOffline, nil)
		return sig.updateRunnerStatus(ctx, runner, status, "no active install process", runner.Org.Name, metadata)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
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

func TestRunnerStatusAfterHealthCheck(t *testing.T) {
	tests := map[string]struct {
		runner         *app.Runner
		status         app.RunnerStatus
		metadata       map[string]any
		expectedStatus app.RunnerStatus
		expectedCount  int
	}{
		"first failed check keeps active runner active": {
			runner:         testRunner(app.RunnerStatusActive),
			status:         app.RunnerStatusOffline,
			expectedStatus: app.RunnerStatusActive,
			expectedCount:  1,
		},
		"second consecutive failed check marks runner offline": {
			runner:         runnerWithHealthCheckFailures(app.RunnerStatusActive, float64(1)),
			status:         app.RunnerStatusOffline,
			expectedStatus: app.RunnerStatusOffline,
			expectedCount:  2,
		},
		"healthy check resets the failure count": {
			runner:         runnerWithHealthCheckFailures(app.RunnerStatusActive, int64(1)),
			status:         app.RunnerStatusActive,
			expectedStatus: app.RunnerStatusActive,
			expectedCount:  0,
		},
		"offline runner remains offline without exceeding threshold": {
			runner:         testRunner(app.RunnerStatusOffline),
			status:         app.RunnerStatusOffline,
			expectedStatus: app.RunnerStatusOffline,
			expectedCount:  2,
		},
		"existing metadata is preserved": {
			runner:         testRunner(app.RunnerStatusActive),
			status:         app.RunnerStatusActive,
			metadata:       map[string]any{"missing_mng_process": true},
			expectedStatus: app.RunnerStatusActive,
			expectedCount:  0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			status, metadata := runnerStatusAfterHealthCheck(tt.runner, tt.status, tt.metadata)

			require.Equal(t, tt.expectedStatus, status)
			require.Equal(t, tt.expectedCount, metadata[app.RunnerHealthCheckConsecutiveFailuresMetadataKey])
			if tt.metadata != nil {
				require.Equal(t, true, metadata["missing_mng_process"])
			}
		})
	}
}

func TestHealthyCheckBreaksFailureStreak(t *testing.T) {
	runner := runnerWithHealthCheckFailures(app.RunnerStatusActive, float64(1))

	status, metadata := runnerStatusAfterHealthCheck(runner, app.RunnerStatusActive, nil)
	require.Equal(t, app.RunnerStatusActive, status)
	require.Equal(t, 0, metadata[app.RunnerHealthCheckConsecutiveFailuresMetadataKey])

	runner.StatusV2.Metadata = metadata
	status, metadata = runnerStatusAfterHealthCheck(runner, app.RunnerStatusOffline, nil)
	require.Equal(t, app.RunnerStatusActive, status)
	require.Equal(t, 1, metadata[app.RunnerHealthCheckConsecutiveFailuresMetadataKey])
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

func runnerWithHealthCheckFailures(status app.RunnerStatus, failures any) *app.Runner {
	runner := testRunner(status)
	runner.StatusV2.Metadata = map[string]any{
		app.RunnerHealthCheckConsecutiveFailuresMetadataKey: failures,
	}
	return runner
}
