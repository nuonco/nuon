package workflows

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

type AppBranchRunStepsTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestAppBranchRunStepsSuite(t *testing.T) {
	suite.Run(t, new(AppBranchRunStepsTestSuite))
}

func (s *AppBranchRunStepsTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.SetWorkerOptions(worker.Options{
		DeadlockDetectionTimeout: time.Minute,
	})
}

// The generator's activities are registered by method name off the Activities
// struct, which needs a database. Stubs registered under the same names keep the
// test to what it is about: the shape of the generated steps.
func (s *AppBranchRunStepsTestSuite) stubActivities(runType app.AppBranchRunType, previews ...*app.AppBranchRunPreview) {
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, req activities.GetAppBranchRunByIDRequest) (*app.AppBranchRun, error) {
			run := &app.AppBranchRun{ID: req.RunID, RunType: runType}
			if len(previews) > 0 {
				run.Preview = previews[0]
			}
			return run, nil
		},
		activity.RegisterOptions{Name: "GetAppBranchRunByID"},
	)
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, configID string) ([]*app.AppBranchInstallGroup, error) {
			return nil, nil
		},
		activity.RegisterOptions{Name: "GetInstallGroupsByConfigID"},
	)
	s.env.RegisterActivityWithOptions(
		func(ctx context.Context, input *activities.GetAppBranchConfigByIDInput) (*activities.GetAppBranchConfigByIDOutput, error) {
			return &activities.GetAppBranchConfigByIDOutput{}, nil
		},
		activity.RegisterOptions{Name: "GetAppBranchConfigByID"},
	)
}

func (s *AppBranchRunStepsTestSuite) generate(metadata map[string]string) *app.GenerateStepsResult {
	hstore := make(map[string]*string, len(metadata))
	for k, v := range metadata {
		val := v
		hstore[k] = &val
	}

	s.env.ExecuteWorkflow(AppBranchRun, &app.Workflow{Metadata: hstore})
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result *app.GenerateStepsResult
	s.NoError(s.env.GetWorkflowResult(&result))
	return result
}

func stepNames(result *app.GenerateStepsResult) []string {
	names := make([]string, 0, len(result.Steps))
	for _, step := range result.Steps {
		names = append(names, step.Name)
	}
	return names
}

func (s *AppBranchRunStepsTestSuite) TestVCSPathFetchesConfig() {
	s.stubActivities(app.AppBranchRunTypeGit)

	result := s.generate(map[string]string{
		"app_branch_id": "branch-1",
		"run_id":        "run-1",
		"config_id":     "config-1",
	})

	s.Equal([]string{"check ignored changes", "fetch commit", "fetch app config", "compute differences", "building components and sandbox"}, stepNames(result))
	s.Equal(app.WorkflowStepExecutionTypeHidden, result.Steps[0].ExecutionType)
}

func (s *AppBranchRunStepsTestSuite) TestPreCompiledConfigSyncsInsteadOfFetching() {
	s.stubActivities(app.AppBranchRunTypeManual)

	result := s.generate(map[string]string{
		"app_branch_id":   "branch-1",
		"run_id":          "run-1",
		"config_id":       "config-1",
		"app_config_id":   "cfg-1",
		"sync_app_config": "true",
	})

	s.Equal([]string{"check ignored changes", "fetch commit (skipped)", "sync app config", "compute differences", "building components and sandbox"}, stepNames(result))
	s.Equal(app.WorkflowStepExecutionTypeSystem, result.Steps[2].ExecutionType)
}

func (s *AppBranchRunStepsTestSuite) TestAlreadySyncedConfigSkipsBothSteps() {
	s.stubActivities(app.AppBranchRunTypeManual)

	result := s.generate(map[string]string{
		"app_branch_id": "branch-1",
		"run_id":        "run-1",
		"config_id":     "config-1",
		"app_config_id": "cfg-1",
		"skip_builds":   "true",
	})

	s.Equal([]string{
		"check ignored changes",
		"fetch commit (skipped)",
		"fetch app config (skipped)",
		"compute differences",
		"building components and sandbox (skipped)",
	}, stepNames(result))
}

func (s *AppBranchRunStepsTestSuite) TestPreviewSetupStepIsHidden() {
	s.stubActivities(
		app.AppBranchRunTypeGitPreview,
		&app.AppBranchRunPreview{Mode: app.AppBranchRunPreviewModeBuildOnly},
	)

	result := s.generate(map[string]string{
		"app_branch_id": "branch-1",
		"run_id":        "run-1",
		"config_id":     "config-1",
	})

	s.Equal("setup preview", result.Steps[1].Name)
	s.Equal(app.WorkflowStepExecutionTypeHidden, result.Steps[1].ExecutionType)
}
