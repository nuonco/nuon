package awaitrunnerhealthy

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func TestStartupWaitsForRequiredProcess(t *testing.T) {
	env, sig, runner := readinessTestEnvironment(t, ModeStartup)
	startedAt := env.Now()
	runner.Status = app.RunnerStatusError
	runner.StatusV2.Status = app.StatusError

	env.OnActivity((*activities.Activities).GetCurrentRunnerProcess, mock.Anything, mock.Anything, mock.Anything).
		Return(testRunnerProcess(app.RunnerProcessStatus(app.StatusPending)), nil).
		Once()
	env.OnActivity((*activities.Activities).GetCurrentRunnerProcess, mock.Anything, mock.Anything, mock.Anything).
		Return(testRunnerProcess(app.RunnerProcessStatusActive), nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.GreaterOrEqual(t, env.Now().Sub(startedAt), 14*time.Second)
	env.AssertExpectations(t)
}

func TestRequireActiveFailsImmediatelyWhenProcessMissing(t *testing.T) {
	env, sig, _ := readinessTestEnvironment(t, ModeRequireActive)
	startedAt := env.Now()

	env.OnActivity((*activities.Activities).GetCurrentRunnerProcess, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, temporal.NewNonRetryableApplicationError("not found", "not found", nil)).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, env.GetWorkflowError(), "runner has no active process")
	require.Equal(t, startedAt, env.Now())
	env.AssertExpectations(t)
}

func TestRequireActiveFailsImmediatelyWhenProcessIsNotActive(t *testing.T) {
	env, sig, _ := readinessTestEnvironment(t, ModeRequireActive)
	startedAt := env.Now()

	env.OnActivity((*activities.Activities).GetCurrentRunnerProcess, mock.Anything, mock.Anything, mock.Anything).
		Return(testRunnerProcess(app.RunnerProcessStatus(app.StatusPending)), nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, env.GetWorkflowError(), "runner process is pending")
	require.Equal(t, startedAt, env.Now())
	env.AssertExpectations(t)
}

func TestDisabledRunnerSkipsProcessReadiness(t *testing.T) {
	env, sig, runner := readinessTestEnvironment(t, ModeRequireActive)
	runner.Status = app.RunnerStatusActive
	runner.StatusV2.Status = app.Status(app.RunnerStatusDisabled)

	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestActiveProcessWinsOverStaleAggregateStatus(t *testing.T) {
	env, sig, runner := readinessTestEnvironment(t, ModeRequireActive)
	runner.Status = app.RunnerStatusOffline
	runner.StatusV2.Status = app.StatusError

	env.OnActivity((*activities.Activities).GetCurrentRunnerProcess, mock.Anything, mock.Anything, mock.Anything).
		Return(testRunnerProcess(app.RunnerProcessStatusActive), nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestExistingHistoryKeepsAggregateFailFastPolicy(t *testing.T) {
	env, sig, runner := readinessTestEnvironment(t, "")
	runner.Status = app.RunnerStatusOffline
	runner.StatusV2.Status = app.Status(app.RunnerStatusOffline)
	env.OnGetVersion(processReadinessPolicyVersion, workflow.DefaultVersion, 1).
		Return(workflow.DefaultVersion).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, env.GetWorkflowError(), "runner is offline")
	env.AssertExpectations(t)
}

func readinessTestEnvironment(t *testing.T, mode Mode) (*testsuite.TestWorkflowEnvironment, *Signal, *app.Runner) {
	t.Helper()

	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})
	env.SetStartTime(time.Date(2026, time.September, 21, 12, 0, 0, 0, time.UTC))

	install := &app.Install{ID: "inl_1", RunnerID: "rnr_1"}
	runner := &app.Runner{
		ID:     install.RunnerID,
		Status: app.RunnerStatusActive,
		StatusV2: app.CompositeStatus{
			Status: app.Status(app.RunnerStatusActive),
		},
		RunnerGroup: app.RunnerGroup{Type: app.RunnerGroupTypeInstall},
	}

	env.OnActivity((*activities.Activities).Get, mock.Anything, mock.Anything, mock.Anything).
		Return(install, nil)
	env.OnActivity((*activities.Activities).GetRunner, mock.Anything, mock.Anything, mock.Anything).
		Return(runner, nil)

	sig := &Signal{InstallID: install.ID, Mode: mode}
	sig.WithParams(&signal.Params{V: validator.New()})
	return env, sig, runner
}

func testRunnerProcess(status app.RunnerProcessStatus) *app.RunnerProcess {
	return &app.RunnerProcess{
		ID:   "rnp_1",
		Type: app.RunnerProcessTypeInstall,
		CompositeStatus: app.CompositeStatus{
			Status: app.Status(status),
		},
	}
}
