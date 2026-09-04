package created

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

type CreatedSignalTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestCreatedSignalSuite(t *testing.T) {
	suite.Run(t, new(CreatedSignalTestSuite))
}

func (s *CreatedSignalTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.SetWorkerOptions(worker.Options{
		DeadlockDetectionTimeout: time.Minute,
	})

	a := &activities.Activities{}
	s.env.RegisterActivityWithOptions(a.AppBranchesGetAppBranchByID, activity.RegisterOptions{Name: "AppBranchesGetAppBranchByID"})
	s.env.RegisterActivityWithOptions(a.TriggerAppBranchRunFromCreated, activity.RegisterOptions{Name: "TriggerAppBranchRunFromCreated"})
}

func (s *CreatedSignalTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *CreatedSignalTestSuite) TestExecuteTriggersFirstRun() {
	sig := &Signal{AppBranchID: "branch-1", AppBranchConfigID: "cfg-1"}

	s.env.OnActivity((*activities.Activities).AppBranchesGetAppBranchByID, mock.Anything, mock.Anything, mock.Anything).Return(
		&app.AppBranch{ID: "branch-1", Name: "main"}, nil)
	s.env.OnActivity((*activities.Activities).TriggerAppBranchRunFromCreated, mock.Anything, mock.Anything, mock.Anything).Return(
		&activities.TriggerAppBranchRunFromCreatedResponse{RunID: "run-1", WorkflowID: "wf-1"}, nil)

	s.env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Execute(ctx)
	})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *CreatedSignalTestSuite) TestExecuteSkipsWhenActivitySkips() {
	sig := &Signal{AppBranchID: "branch-1", AppBranchConfigID: "cfg-1"}

	s.env.OnActivity((*activities.Activities).AppBranchesGetAppBranchByID, mock.Anything, mock.Anything, mock.Anything).Return(
		&app.AppBranch{ID: "branch-1", Name: "default"}, nil)
	s.env.OnActivity((*activities.Activities).TriggerAppBranchRunFromCreated, mock.Anything, mock.Anything, mock.Anything).Return(
		&activities.TriggerAppBranchRunFromCreatedResponse{Skipped: true, Reason: "no_vcs"}, nil)

	s.env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Execute(ctx)
	})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}
