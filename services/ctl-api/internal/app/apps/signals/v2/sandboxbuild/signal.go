package sandboxbuild

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	workerplan "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	jobpkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

// SignalType is the type for direct (non-branch) sandbox build signals.
const SignalType signal.SignalType = "app-sandbox-build"

// Signal triggers a sandbox build for a given app config.
// If AppSandboxBuildID is set, the existing build record is used; otherwise a new one is created.
type Signal struct {
	signal.Hooks
	AppConfigID       string `json:"app_config_id" validate:"required"`
	AppSandboxBuildID string `json:"app_sandbox_build_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.AppConfigID == "" {
		return fmt.Errorf("app_config_id is required")
	}

	_, err := activities.AwaitGetAppConfigByIDByAppConfigID(ctx, s.AppConfigID)
	if err != nil {
		return fmt.Errorf("app config not found: %w", err)
	}

	return nil
}

// getOrCreateBuild returns the existing build if AppSandboxBuildID is set, otherwise creates a new one.
func (s *Signal) getOrCreateBuild(ctx workflow.Context, appConfig *app.AppConfig, sandboxConfig *app.AppSandboxConfig) (*app.AppSandboxBuild, error) {
	if s.AppSandboxBuildID != "" {
		build, err := activities.AwaitGetAppSandboxBuildByIDByBuildID(ctx, s.AppSandboxBuildID)
		if err != nil {
			return nil, fmt.Errorf("unable to get sandbox build %s: %w", s.AppSandboxBuildID, err)
		}
		return build, nil
	}

	build, err := activities.AwaitCreateSandboxBuild(ctx, activities.CreateSandboxBuildRequest{
		AppID:              appConfig.AppID,
		AppConfigID:        s.AppConfigID,
		AppSandboxConfigID: sandboxConfig.ID,
		OrgID:              appConfig.OrgID,
		CreatedByID:        appConfig.CreatedByID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create sandbox build: %w", err)
	}
	return build, nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Resolve app config → app ID
	appConfig, err := activities.AwaitGetAppConfigByIDByAppConfigID(ctx, s.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	// Get the sandbox config for this app
	sandboxConfig, err := activities.AwaitGetLatestAppSandboxConfigByAppID(ctx, appConfig.AppID)
	if err != nil {
		return fmt.Errorf("unable to get sandbox config: %w", err)
	}

	// Get org runner
	runner, err := activities.AwaitGetOrgRunner(ctx, activities.GetOrgRunnerRequest{OrgID: appConfig.OrgID})
	if err != nil {
		return fmt.Errorf("unable to get org runner: %w", err)
	}

	// Get or create the sandbox build record
	build, err := s.getOrCreateBuild(ctx, appConfig, sandboxConfig)
	if err != nil {
		return err
	}

	l.Info("sandbox build started", "build_id", build.ID)

	// Create log stream
	logStreamID := ""
	logStream, logStreamErr := activities.AwaitCreateSandboxBuildLogStream(ctx, activities.CreateSandboxBuildLogStreamRequest{
		AppSandboxBuildID: build.ID,
		OrgID:             build.OrgID,
		CreatedByID:       build.CreatedByID,
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

	// Create runner job
	runnerJob, err := activities.AwaitCreateSandboxBuildJob(ctx, activities.CreateSandboxBuildJobRequest{
		BuildID:     build.ID,
		RunnerID:    runner.ID,
		LogStreamID: logStreamID,
	})
	if err != nil {
		updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create sandbox build job: %w", err)
	}

	// Build the plan via child workflow
	buildPlan, err := workerplan.AwaitCreateSandboxBuildPlan(ctx, &workerplan.CreateSandboxBuildPlanRequest{
		AppSandboxBuildID: build.ID,
		WorkflowID:        fmt.Sprintf("%s-create-sandbox-build-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to create build plan")
		return fmt.Errorf("unable to create sandbox build plan: %w", err)
	}

	planJSON, err := json.Marshal(buildPlan)
	if err != nil {
		updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to marshal build plan")
		return fmt.Errorf("unable to marshal plan: %w", err)
	}

	if err := activities.AwaitSaveSandboxBuildPlan(ctx, activities.SaveSandboxBuildPlanRequest{
		JobID:         runnerJob.ID,
		CompositePlan: plantypes.CompositePlan{BuildPlan: buildPlan},
		PlanJSON:      string(planJSON),
	}); err != nil {
		updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to save build plan")
		return fmt.Errorf("unable to save sandbox build plan: %w", err)
	}

	updateStatus(ctx, build.ID, app.AppSandboxBuildStatusPlanning, "planning sandbox build")

	// Execute the runner job
	_, err = jobpkg.AwaitExecuteJob(ctx, &jobpkg.ExecuteJobRequest{
		RunnerID:   runner.ID,
		JobID:      runnerJob.ID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", build.ID, runnerJob.ID),
	})
	if err != nil {
		updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "sandbox build job failed")
		return fmt.Errorf("sandbox build job failed: %w", err)
	}

	updateStatus(ctx, build.ID, app.AppSandboxBuildStatusActive, "sandbox build completed")
	l.Info("sandbox build completed successfully", "build_id", build.ID)
	return nil
}

func updateStatus(ctx workflow.Context, buildID string, status app.AppSandboxBuildStatus, description string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateSandboxBuildStatus(ctx, activities.UpdateSandboxBuildStatusRequest{
		BuildID:           buildID,
		Status:            status,
		StatusDescription: description,
	}); err != nil {
		l.Error("unable to update sandbox build status", "error", err, "build_id", buildID)
	}
}
