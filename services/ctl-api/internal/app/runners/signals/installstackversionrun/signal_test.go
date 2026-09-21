package installstackversionrun

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type InstallStackVersionRunSignalTestSuite struct {
	suite.Suite
}

func TestInstallStackVersionRunSignalSuite(t *testing.T) {
	suite.Run(t, new(InstallStackVersionRunSignalTestSuite))
}

func (s *InstallStackVersionRunSignalTestSuite) TestInstallStackVersionRunSignalExecutesSuccessfully() {
	env, sig := s.testEnvironment(app.RunnerStatusAwaitingInstallStackRun)
	env.OnActivity((*activities.Activities).GetStackRunRunnerDisabled, mock.Anything, mock.Anything, mock.Anything).
		Return(false, nil).
		Once()
	env.OnActivity((*activities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })

	require.NoError(s.T(), env.GetWorkflowError())
	env.AssertExpectations(s.T())
}

func (s *InstallStackVersionRunSignalTestSuite) TestInstallStackVersionRunSignalSkipsIfNotAwaitingRun() {
	env, sig := s.testEnvironment(app.RunnerStatusActive)

	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })

	require.NoError(s.T(), env.GetWorkflowError())
	env.AssertExpectations(s.T())
}

func (s *InstallStackVersionRunSignalTestSuite) TestInstallStackVersionRunSignalValidationFails() {
	sig := &Signal{}
	require.EqualError(s.T(), sig.Validate(nil), "runner_id is required")
	sig.RunnerID = "rnr_1"
	require.EqualError(s.T(), sig.Validate(nil), "install_stack_version_run_id is required")
}

func (s *InstallStackVersionRunSignalTestSuite) TestInstallStackVersionRunSignalPropagatesWriteFailure() {
	env, sig := s.testEnvironment(app.RunnerStatusAwaitingInstallStackRun)
	env.OnActivity((*activities.Activities).GetStackRunRunnerDisabled, mock.Anything, mock.Anything, mock.Anything).
		Return(false, nil).
		Once()
	env.OnActivity((*activities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(temporal.NewNonRetryableApplicationError("write failed", "test", nil)).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })

	require.ErrorContains(s.T(), env.GetWorkflowError(), "write failed")
	env.AssertExpectations(s.T())
}

func (s *InstallStackVersionRunSignalTestSuite) TestExistingHistoryKeepsSplitStatusWrite() {
	env, sig := s.testEnvironment(app.RunnerStatusAwaitingInstallStackRun)
	env.OnGetVersion(atomicRunnerStatusVersion, workflow.DefaultVersion, 1).
		Return(workflow.DefaultVersion).
		Once()
	env.OnActivity((*activities.Activities).GetStackRunRunnerDisabled, mock.Anything, mock.Anything, mock.Anything).
		Return(false, nil).
		Once()
	env.OnActivity((*activities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()
	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })

	require.NoError(s.T(), env.GetWorkflowError())
	env.AssertExpectations(s.T())
}

func (s *InstallStackVersionRunSignalTestSuite) testEnvironment(status app.RunnerStatus) (*testsuite.TestWorkflowEnvironment, *Signal) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

	sig := &Signal{RunnerID: "rnr_1", InstallStackVersionRunID: "isvr_1"}
	runner := &app.Runner{
		ID:     sig.RunnerID,
		Status: status,
		StatusV2: app.CompositeStatus{
			Status: app.Status(status),
		},
	}
	env.OnActivity((*activities.Activities).Get, mock.Anything, mock.Anything, mock.Anything).
		Return(runner, nil).
		Once()
	return env, sig
}
