package worker

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	workerplan "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/plan"
	jobpkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

// @temporal-gen-v2 workflow
// @execution-timeout 30m
func (w *Workflows) BuildSandbox(ctx workflow.Context, sreq signals.RequestSignal) error {
	l := workflow.GetLogger(ctx)

	// Get the pre-created sandbox build record
	build, err := branchactivities.AwaitGetAppSandboxBuildByIDByBuildID(ctx, sreq.AppSandboxBuildID)
	if err != nil {
		return fmt.Errorf("unable to get sandbox build: %w", err)
	}

	// Get org runner
	runner, err := branchactivities.AwaitGetOrgRunner(ctx, branchactivities.GetOrgRunnerRequest{OrgID: build.OrgID})
	if err != nil {
		updateSandboxBuildStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to get org runner")
		return fmt.Errorf("unable to get org runner: %w", err)
	}

	// Create log stream
	logStreamID := ""
	logStream, logStreamErr := branchactivities.AwaitCreateSandboxBuildLogStream(ctx, branchactivities.CreateSandboxBuildLogStreamRequest{
		AppSandboxBuildID: build.ID,
		OrgID:             build.OrgID,
		CreatedByID:       build.CreatedByID,
	})
	if logStreamErr != nil {
		l.Warn("unable to create log stream for sandbox build, continuing without it", "error", logStreamErr)
	} else if logStream != nil {
		logStreamID = logStream.ID
		defer func() {
			_ = branchactivities.AwaitCloseLogStream(ctx, branchactivities.CloseLogStreamRequest{
				LogStreamID: logStreamID,
			})
		}()
	}

	// Create runner job
	runnerJob, err := branchactivities.AwaitCreateSandboxBuildJob(ctx, branchactivities.CreateSandboxBuildJobRequest{
		BuildID:     build.ID,
		RunnerID:    runner.ID,
		LogStreamID: logStreamID,
	})
	if err != nil {
		updateSandboxBuildStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create sandbox build job: %w", err)
	}

	// Build the plan via child workflow
	buildPlan, err := workerplan.AwaitCreateSandboxBuildPlan(ctx, &workerplan.CreateSandboxBuildPlanRequest{
		AppSandboxBuildID: build.ID,
		WorkflowID:        fmt.Sprintf("%s-create-sandbox-build-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		updateSandboxBuildStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to create build plan")
		return fmt.Errorf("unable to create sandbox build plan: %w", err)
	}

	planJSON, err := json.Marshal(buildPlan)
	if err != nil {
		updateSandboxBuildStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to marshal build plan")
		return fmt.Errorf("unable to marshal plan: %w", err)
	}

	if err := branchactivities.AwaitSaveSandboxBuildPlan(ctx, branchactivities.SaveSandboxBuildPlanRequest{
		JobID:         runnerJob.ID,
		CompositePlan: plantypes.CompositePlan{BuildPlan: buildPlan},
		PlanJSON:      string(planJSON),
	}); err != nil {
		updateSandboxBuildStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to save build plan")
		return fmt.Errorf("unable to save sandbox build plan: %w", err)
	}

	updateSandboxBuildStatus(ctx, build.ID, app.AppSandboxBuildStatusPlanning, "planning sandbox build")

	// Execute the runner job
	_, err = jobpkg.AwaitExecuteJob(ctx, &jobpkg.ExecuteJobRequest{
		RunnerID:   runner.ID,
		JobID:      runnerJob.ID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", build.ID, runnerJob.ID),
	})
	if err != nil {
		updateSandboxBuildStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "sandbox build job failed")
		return fmt.Errorf("sandbox build job failed: %w", err)
	}

	updateSandboxBuildStatus(ctx, build.ID, app.AppSandboxBuildStatusActive, "sandbox build completed")
	l.Info("sandbox build completed successfully", "build_id", build.ID)
	return nil
}

func updateSandboxBuildStatus(ctx workflow.Context, buildID string, status app.AppSandboxBuildStatus, description string) {
	l := workflow.GetLogger(ctx)
	if err := branchactivities.AwaitUpdateSandboxBuildStatus(ctx, branchactivities.UpdateSandboxBuildStatusRequest{
		BuildID:           buildID,
		Status:            status,
		StatusDescription: description,
	}); err != nil {
		l.Error("unable to update sandbox build status", "error", err, "build_id", buildID)
	}
}
