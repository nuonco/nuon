package sandboxbuild

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	jobpkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Get the run with VCS commit
	run, err := activities.AwaitGetAppBranchRunWithCommitByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if run.AppConfigID == "" {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

	// Get the app config (with App preloaded) to get AppID
	appConfig, err := activities.AwaitGetAppConfigByIDByAppConfigID(ctx, run.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	// Get the sandbox config for this app — if not found, skip gracefully
	sandboxConfig, err := activities.AwaitGetLatestAppSandboxConfigByAppID(ctx, appConfig.AppID)
	if err != nil {
		l.Info("no sandbox config found for app, skipping sandbox build", "app_id", appConfig.AppID)
		return nil
	}

	// Get the org runner
	runner, err := activities.AwaitGetOrgRunner(ctx, activities.GetOrgRunnerRequest{OrgID: run.OrgID})
	if err != nil {
		return fmt.Errorf("unable to get org runner: %w", err)
	}

	// Resolve git source for the sandbox build
	gitSource, err := activities.AwaitGetSandboxBuildGitSource(ctx, activities.GetSandboxBuildGitSourceRequest{
		SandboxConfigID: sandboxConfig.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to get sandbox build git source: %w", err)
	}

	l.Info("hello world test")

	// If sandbox config shares the same VCS config as the branch run's commit, pin to that specific SHA
	if run.VCSConnectionCommit != nil {
		var sandboxVCSConfigID string
		if sandboxConfig.ConnectedGithubVCSConfig != nil {
			sandboxVCSConfigID = sandboxConfig.ConnectedGithubVCSConfig.ID
		} else if sandboxConfig.PublicGitVCSConfig != nil {
			sandboxVCSConfigID = sandboxConfig.PublicGitVCSConfig.ID
		}
		if sandboxVCSConfigID != "" && sandboxVCSConfigID == run.VCSConnectionCommit.OwnerID {
			gitSource.Ref = run.VCSConnectionCommit.SHA
		}
	}

	// Resolve VCS commit ID for the sandbox build record
	var commitID *string
	if run.VCSConnectionCommit != nil {
		commitID = &run.VCSConnectionCommit.ID
	}

	// Create the sandbox build record
	build, err := activities.AwaitCreateSandboxBuild(ctx, activities.CreateSandboxBuildRequest{
		AppID:                 appConfig.AppID,
		AppConfigID:           run.AppConfigID,
		AppSandboxConfigID:    sandboxConfig.ID,
		OrgID:                 run.OrgID,
		CreatedByID:           run.CreatedByID,
		VCSConnectionCommitID: commitID,
	})
	if err != nil {
		return fmt.Errorf("unable to create sandbox build: %w", err)
	}

	l.Info("sandbox build created", "build_id", build.ID)

	// Create a log stream for the sandbox build
	logStreamID := ""
	logStream, logStreamErr := activities.AwaitCreateSandboxBuildLogStream(ctx, activities.CreateSandboxBuildLogStreamRequest{
		AppSandboxBuildID: build.ID,
		OrgID:             run.OrgID,
		CreatedByID:       run.CreatedByID,
	})
	if logStreamErr != nil {
		l.Warn("unable to create log stream for sandbox build, continuing without it", "error", logStreamErr)
	} else if logStream != nil {
		logStreamID = logStream.ID
		defer func() {
			_ = activities.AwaitCloseLogStream(ctx, activities.CloseLogStreamRequest{
				LogStreamID: logStreamID,
			})
		}()
	}

	// Create the runner job
	runnerJob, err := activities.AwaitCreateSandboxBuildJob(ctx, activities.CreateSandboxBuildJobRequest{
		BuildID:     build.ID,
		RunnerID:    runner.ID,
		LogStreamID: logStreamID,
	})
	if err != nil {
		s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create sandbox build job: %w", err)
	}

	// Build the composite plan
	compositePlan := plantypes.CompositePlan{
		BuildPlan: &plantypes.BuildPlan{
			Src: gitSource,
			TerraformBuildPlan: &plantypes.TerraformBuildPlan{
				Labels: map[string]string{
					"app_id":               appConfig.AppID,
					"app_sandbox_build_id": build.ID,
				},
			},
		},
	}
	planJSON, err := json.Marshal(compositePlan)
	if err != nil {
		s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to create build plan")
		return fmt.Errorf("unable to marshal plan: %w", err)
	}

	if err := activities.AwaitSaveSandboxBuildPlan(ctx, activities.SaveSandboxBuildPlanRequest{
		JobID:         runnerJob.ID,
		CompositePlan: compositePlan,
		PlanJSON:      string(planJSON),
	}); err != nil {
		s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to save build plan")
		return fmt.Errorf("unable to save sandbox build plan: %w", err)
	}

	s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusPlanning, "planning sandbox build")

	// Execute the runner job
	_, err = jobpkg.AwaitExecuteJob(ctx, &jobpkg.ExecuteJobRequest{
		RunnerID:   runner.ID,
		JobID:      runnerJob.ID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", build.ID, runnerJob.ID),
	})
	if err != nil {
		s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "sandbox build job failed")
		return fmt.Errorf("sandbox build job failed: %w", err)
	}

	s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusActive, "sandbox build completed")
	l.Info("sandbox build completed successfully", "build_id", build.ID)
	return nil
}

func (s *Signal) updateStatus(ctx workflow.Context, buildID string, status app.AppSandboxBuildStatus, description string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateSandboxBuildStatus(ctx, activities.UpdateSandboxBuildStatusRequest{
		BuildID:           buildID,
		Status:            status,
		StatusDescription: description,
	}); err != nil {
		l.Error("unable to update sandbox build status", "error", err, "build_id", buildID)
	}
}
