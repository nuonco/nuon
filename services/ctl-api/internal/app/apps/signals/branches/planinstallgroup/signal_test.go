package planinstallgroup

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

type PlanInstallGroupTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestPlanInstallGroupSuite(t *testing.T) {
	suite.Run(t, new(PlanInstallGroupTestSuite))
}

func (s *PlanInstallGroupTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	// Give the deadlock detector generous headroom so loaded CI runners don't
	// false-positive on the 1s default while the workflow goroutine runs.
	s.env.SetWorkerOptions(worker.Options{
		DeadlockDetectionTimeout: time.Minute,
	})

	// Register activities with string-based names so the test env can match them
	// when called via method references with primitive (non-struct) arguments.
	a := &activities.Activities{}
	s.env.RegisterActivityWithOptions(a.GetInstallGroupByID, activity.RegisterOptions{Name: "GetInstallGroupByID"})
	s.env.RegisterActivityWithOptions(a.GetInstall, activity.RegisterOptions{Name: "GetInstall"})
}

func (s *PlanInstallGroupTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *PlanInstallGroupTestSuite) TestEmptyInstallGroup() {
	sig := &Signal{
		InstallGroupID: "group-1",
		AppBranchID:    "branch-1",
		RunID:          "run-1",
		StepID:         "step-1",
	}

	s.env.OnActivity((*activities.Activities).GetAppBranchRunByID, mock.Anything, mock.Anything, mock.Anything).Return(
		&app.AppBranchRun{
			ID:          "run-1",
			AppConfigID: "cfg-1",
		}, nil)

	s.env.OnActivity("GetInstallGroupByID", mock.Anything, mock.Anything).Return(
		&app.AppBranchInstallGroup{
			ID:         "group-1",
			Name:       "prod",
			InstallIDs: []string{},
		}, nil)

	s.env.ExecuteWorkflow(sig.Execute)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *PlanInstallGroupTestSuite) TestNoStepIDSkipsApproval() {
	sig := &Signal{
		InstallGroupID: "group-1",
		AppBranchID:    "branch-1",
		RunID:          "run-1",
		StepID:         "",
	}

	s.env.OnActivity((*activities.Activities).GetAppBranchRunByID, mock.Anything, mock.Anything, mock.Anything).Return(
		&app.AppBranchRun{
			ID:          "run-1",
			AppConfigID: "cfg-1",
		}, nil)

	s.env.OnActivity("GetInstallGroupByID", mock.Anything, mock.Anything).Return(
		&app.AppBranchInstallGroup{
			ID:         "group-1",
			Name:       "prod",
			InstallIDs: []string{"install-1"},
		}, nil)
	s.env.OnActivity((*activities.Activities).ResolveInstallGroupInstalls, mock.Anything, mock.Anything, mock.Anything).Return(
		&activities.ResolveInstallGroupInstallsOutput{InstallIDs: []string{"install-1"}}, nil)

	s.env.OnActivity("GetInstall", mock.Anything, mock.Anything).Return(
		&app.Install{
			ID:          "install-1",
			Name:        "test-install",
			AppConfigID: "old-cfg-1",
			Labeled:     labels.Labeled{Labels: labels.Labels{"env": "prod"}},
		}, nil)

	s.env.OnActivity((*activities.Activities).ComputeInstallConfigDiff, mock.Anything, mock.Anything, mock.Anything).Return(
		&activities.ComputeInstallConfigDiffOutput{
			Diff: &app.InstallConfigDiff{
				Added:   []app.ComponentDiffEntry{{ComponentID: "comp-1", ComponentName: "api"}},
				Changed: []app.ComponentDiffEntry{},
				Removed: []app.ComponentDiffEntry{},
			},
		}, nil)

	s.env.ExecuteWorkflow(sig.Execute)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *PlanInstallGroupTestSuite) TestLabelSelectorResolution() {
	sig := &Signal{
		InstallGroupID: "group-1",
		AppBranchID:    "branch-1",
		RunID:          "run-1",
		StepID:         "",
	}

	s.env.OnActivity((*activities.Activities).GetAppBranchRunByID, mock.Anything, mock.Anything, mock.Anything).Return(
		&app.AppBranchRun{
			ID:          "run-1",
			AppConfigID: "cfg-1",
		}, nil)

	selector := &labels.Selector{
		MatchLabels: labels.Labels{"env": "staging"},
	}
	s.env.OnActivity("GetInstallGroupByID", mock.Anything, mock.Anything).Return(
		&app.AppBranchInstallGroup{
			ID:            "group-1",
			Name:          "staging",
			LabelSelector: selector,
		}, nil)

	s.env.OnActivity((*activities.Activities).AppBranchesGetAppBranchByID, mock.Anything, mock.Anything, mock.Anything).Return(
		&app.AppBranch{
			ID:    "branch-1",
			AppID: "app-1",
		}, nil)

	s.env.OnActivity((*activities.Activities).ResolveInstallGroupInstalls, mock.Anything, mock.Anything, mock.Anything).Return(
		&activities.ResolveInstallGroupInstallsOutput{
			InstallIDs: []string{"install-staging-1"},
		}, nil)

	s.env.OnActivity("GetInstall", mock.Anything, mock.Anything).Return(
		&app.Install{
			ID:          "install-staging-1",
			Name:        "staging-install",
			AppConfigID: "old-cfg",
			Labeled:     labels.Labeled{Labels: labels.Labels{"env": "staging"}},
		}, nil)

	s.env.OnActivity((*activities.Activities).ComputeInstallConfigDiff, mock.Anything, mock.Anything, mock.Anything).Return(
		&activities.ComputeInstallConfigDiffOutput{
			Diff: &app.InstallConfigDiff{
				Changed: []app.ComponentDiffEntry{{ComponentID: "comp-1", ComponentName: "api"}},
			},
		}, nil)

	s.env.ExecuteWorkflow(sig.Execute)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}
