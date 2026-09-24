package sandboxbuild

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/controlplanejob"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const sourceAfterBuildVersion = "app-branch-sandbox-build-source-after-build-v1"

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// The commit is optional here — it only pins a SHA below. A run triggered with
	// a pre-existing app config (the CLI's branch-targeted sync) skips fetch-commit
	// and has none, so requiring it wedges this step forever.
	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if run.AppConfigID == "" {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

	sourceAfterBuild := workflow.GetVersion(ctx, sourceAfterBuildVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion

	var (
		sandboxConfig *app.AppSandboxConfig
		source        *activities.ResolveSandboxBuildSourceOutput
	)
	if sourceAfterBuild {
		sandbox, err := activities.AwaitGetSandboxBuildConfig(ctx, &activities.GetSandboxBuildConfigInput{
			AppConfigID: run.AppConfigID,
		})
		if err != nil {
			return fmt.Errorf("unable to get sandbox build config: %w", err)
		}
		if sandbox == nil || sandbox.Skipped || sandbox.SandboxConfig == nil {
			l.Info("no sandbox config found for app config, skipping sandbox build", "app_config_id", run.AppConfigID)
			return nil
		}
		sandboxConfig = sandbox.SandboxConfig
	} else {
		source, err = activities.AwaitResolveSandboxBuildSource(ctx, &activities.ResolveSandboxBuildSourceInput{
			AppConfigID: run.AppConfigID,
			RunID:       s.RunID,
		})
		if err != nil {
			return fmt.Errorf("unable to resolve sandbox build source: %w", err)
		}
		if source == nil || source.Skipped || source.SandboxConfig == nil {
			l.Info("no sandbox config found for app config, skipping sandbox build", "app_config_id", run.AppConfigID)
			return nil
		}
		sandboxConfig = source.SandboxConfig
	}

	appConfig, err := activities.AwaitGetAppConfigByIDByAppConfigID(ctx, run.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	createReq := activities.CreateSandboxBuildRequest{
		AppID:              appConfig.AppID,
		AppConfigID:        run.AppConfigID,
		AppSandboxConfigID: sandboxConfig.ID,
		OrgID:              run.OrgID,
		CreatedByID:        run.CreatedByID,
		AppBranchRunID:     s.RunID,
	}
	if !sourceAfterBuild {
		createReq.VCSConnectionCommitID = source.VCSConnectionCommitID
	}
	build, err := activities.AwaitCreateSandboxBuild(ctx, createReq)
	if err != nil {
		return fmt.Errorf("unable to create sandbox build: %w", err)
	}

	l.Info("sandbox build created", "build_id", build.ID)
	if s.StepID != "" {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.StepID,
			Status: app.CompositeStatus{
				Status:                 app.StatusInProgress,
				StatusHumanDescription: "building sandbox",
				Metadata: map[string]any{
					"sandbox_build_id": build.ID,
				},
			},
		}); err != nil {
			l.Warn("unable to add sandbox build to step metadata", "error", err, "build_id", build.ID)
		}
	}

	if sourceAfterBuild {
		source, err = activities.AwaitResolveSandboxBuildSource(ctx, &activities.ResolveSandboxBuildSourceInput{
			AppConfigID: run.AppConfigID,
			RunID:       s.RunID,
			BuildID:     build.ID,
		})
		if err != nil {
			s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, activities.SandboxSourceFailureDescription(err))
			return fmt.Errorf("unable to resolve sandbox build source: %w", err)
		}
	}
	if source == nil || source.GitSource == nil {
		s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to resolve sandbox source")
		return fmt.Errorf("sandbox build source resolved to nothing for app config %s", run.AppConfigID)
	}

	gitSource := source.GitSource

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
		LogStreamID: logStreamID,
	})
	if err != nil {
		s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create sandbox build job: %w", err)
	}

	dstCfg, err := sharedactivities.AwaitGetSandboxBuildOCIRegistry(ctx, sharedactivities.GetSandboxBuildOCIRegistryRequest{
		AppID: appConfig.AppID,
	})
	if err != nil {
		s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to get OCI registry")
		return fmt.Errorf("unable to get OCI registry for sandbox build: %w", err)
	}

	tfPlan := &plantypes.TerraformBuildPlan{
		Labels: map[string]string{
			"app_id":               appConfig.AppID,
			"app_sandbox_build_id": build.ID,
		},
	}
	mirrorEnabled, err := activities.AwaitOrgHasFeature(ctx, activities.OrgHasFeatureRequest{
		OrgID:   run.OrgID,
		Feature: string(app.OrgFeatureTerraformProviderMirror),
	})
	if err != nil {
		s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusError, "unable to check terraform provider mirror feature flag")
		return fmt.Errorf("unable to check terraform provider mirror feature flag: %w", err)
	}
	if mirrorEnabled {
		tfPlan.VendorProviders = true
		tfPlan.TerraformVersion = sandboxConfig.TerraformVersion
	}

	isSandboxOrg, err := activities.AwaitIsOrgSandboxMode(ctx, activities.IsOrgSandboxModeRequest{
		OrgID: run.OrgID,
	})
	if err != nil {
		l.Warn("unable to check org sandbox mode, continuing without it", "error", err)
	}

	compositePlan := plantypes.CompositePlan{
		BuildPlan: &plantypes.BuildPlan{
			Src:                gitSource,
			Dst:                dstCfg,
			DstTag:             build.ID,
			TerraformBuildPlan: tfPlan,
		},
	}
	if isSandboxOrg {
		compositePlan.BuildPlan.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
		}
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
	s.updateStatus(ctx, build.ID, app.AppSandboxBuildStatusBuilding, "building sandbox")
	err = controlplanejob.AwaitExecuteControlPlaneJob(ctx, &controlplanejob.ExecuteRequest{JobID: runnerJob.ID}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("control-plane-%s-execute-job-%s", build.ID, runnerJob.ID),
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
