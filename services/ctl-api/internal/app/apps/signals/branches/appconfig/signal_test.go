package appconfig

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

type AppConfigSignalTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestAppConfigSignalSuite(t *testing.T) {
	suite.Run(t, new(AppConfigSignalTestSuite))
}

func (s *AppConfigSignalTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	// Give the deadlock detector generous headroom so loaded CI runners don't
	// false-positive on the 1s default while the workflow goroutine runs.
	s.env.SetWorkerOptions(worker.Options{
		DeadlockDetectionTimeout: time.Minute,
	})
}

func (s *AppConfigSignalTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *AppConfigSignalTestSuite) mockBranchAndRun() {
	s.env.OnActivity((*activities.Activities).GetAppBranchRunWithCommit, mock.Anything, mock.Anything, mock.Anything).Return(
		&app.AppBranchRun{
			ID: "run-1",
			VCSConnectionCommit: &app.VCSConnectionCommit{
				SHA: "abc123",
			},
		}, nil)

	s.env.OnActivity((*activities.Activities).CreateLogStream, mock.Anything, mock.Anything, mock.Anything).Return(
		(*app.LogStream)(nil), fmt.Errorf("no log stream"))

	s.env.OnActivity((*activities.Activities).AppBranchesGetAppBranchByID, mock.Anything, mock.Anything, mock.Anything).Return(
		&app.AppBranch{
			ID:    "branch-1",
			AppID: "app-1",
			OrgID: "org-1",
			Configs: []app.AppBranchConfig{
				{
					ID: "config-1",
					ConnectedGithubVCSConfig: &app.ConnectedGithubVCSConfig{
						ID:     "vcs-1",
						Repo:   "owner/repo",
						Branch: "main",
					},
				},
			},
		}, nil)
}

// mockPreCompiledBranchAndRun stands up a branch with no VCS config at all, the
// shape the default app branch has.
func (s *AppConfigSignalTestSuite) mockPreCompiledBranchAndRun() {
	s.env.OnActivity((*activities.Activities).GetAppBranchRunByID, mock.Anything, mock.Anything, mock.Anything).Return(
		&app.AppBranchRun{ID: "run-1"}, nil)

	s.env.OnActivity((*activities.Activities).CreateLogStream, mock.Anything, mock.Anything, mock.Anything).Return(
		(*app.LogStream)(nil), fmt.Errorf("no log stream"))

	s.env.OnActivity((*activities.Activities).AppBranchesGetAppBranchByID, mock.Anything, mock.Anything, mock.Anything).Return(
		&app.AppBranch{
			ID:      "branch-1",
			AppID:   "app-1",
			OrgID:   "org-1",
			Configs: []app.AppBranchConfig{{ID: "config-1"}},
		}, nil)
}

// A pre-compiled config syncs without cloning: CloneRepo and
// FetchIntermediateConfig are left unmocked, so the workflow fails if it reaches
// them.
func (s *AppConfigSignalTestSuite) TestPreCompiledConfigSyncsWithoutClone() {
	sig := &Signal{
		AppBranchID: "branch-1",
		RunID:       "run-1",
		AppConfigID: "cfg-1",
	}

	s.mockPreCompiledBranchAndRun()

	s.env.OnActivity((*activities.Activities).SyncAppConfig, mock.Anything, mock.Anything, mock.Anything).Return(
		&activities.SyncAppConfigOutput{
			AppConfigID:  "cfg-1",
			ComponentIDs: []string{"cmp-1"},
		}, nil)

	s.env.OnActivity((*activities.Activities).UpdateAppBranchConfigIDs, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity((*activities.Activities).UpdateAppBranchRunAppConfig, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(sig.Execute)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *AppConfigSignalTestSuite) TestPreCompiledConfigSyncError() {
	sig := &Signal{
		AppBranchID: "branch-1",
		RunID:       "run-1",
		AppConfigID: "cfg-1",
	}

	s.mockPreCompiledBranchAndRun()

	s.env.OnActivity((*activities.Activities).SyncAppConfig, mock.Anything, mock.Anything, mock.Anything).Return(
		(*activities.SyncAppConfigOutput)(nil), fmt.Errorf("component validation failed"))

	s.env.ExecuteWorkflow(sig.Execute)

	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)
	s.Contains(err.Error(), "unable to sync app config")
}

func (s *AppConfigSignalTestSuite) TestFetchIntermediateConfigError() {
	sig := &Signal{
		AppBranchID: "branch-1",
		RunID:       "run-1",
		StepID:      "step-1",
	}

	s.mockBranchAndRun()

	s.env.OnActivity((*activities.Activities).CloneRepo, mock.Anything, mock.Anything, mock.Anything).Return(
		&activities.CloneRepoResult{SourceDir: "/tmp/repo"}, nil)

	s.env.OnActivity((*activities.Activities).FetchIntermediateConfig, mock.Anything, mock.Anything, mock.Anything).Return(
		(*config.AppConfig)(nil), fmt.Errorf("invalid nuon.toml: missing version field"))

	s.env.ExecuteWorkflow(sig.Execute)

	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)
	s.Contains(err.Error(), "unable to fetch intermediate config")
}

func (s *AppConfigSignalTestSuite) TestSyncAppConfigError() {
	sig := &Signal{
		AppBranchID: "branch-1",
		RunID:       "run-1",
		StepID:      "step-1",
	}

	s.mockBranchAndRun()

	s.env.OnActivity((*activities.Activities).CloneRepo, mock.Anything, mock.Anything, mock.Anything).Return(
		&activities.CloneRepoResult{SourceDir: "/tmp/repo"}, nil)

	s.env.OnActivity((*activities.Activities).FetchIntermediateConfig, mock.Anything, mock.Anything, mock.Anything).Return(
		&config.AppConfig{Version: "v1"}, nil)

	s.env.OnActivity((*activities.Activities).CreateAppConfig, mock.Anything, mock.Anything, mock.Anything).Return(
		&activities.CreateAppConfigOutput{AppConfigID: "cfg-1"}, nil)

	s.env.OnActivity((*activities.Activities).SyncAppConfig, mock.Anything, mock.Anything, mock.Anything).Return(
		(*activities.SyncAppConfigOutput)(nil), fmt.Errorf("component validation failed"))

	s.env.ExecuteWorkflow(sig.Execute)

	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)
	s.Contains(err.Error(), "unable to sync app config")
}
