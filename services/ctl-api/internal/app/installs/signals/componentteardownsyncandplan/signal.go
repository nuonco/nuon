package componentteardownsyncandplan

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/imagesync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/workflowstepapprovalrequest"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "component-teardown-sync-and-plan"

const planCompositeErrorVersion = "teardown-sync-and-plan-composite-error-v1"

type Signal struct {
	signal.LifecycleBase

	InstallComponentID string
	InstallID          string
	ComponentID        string
	WorkflowStepID     string
	FlowStepID         string
	FlowID             string
	SandboxMode        bool
	Role               string

	runnerJobID string
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID = stepID
	s.FlowStepID = stepID
	s.FlowID = flowID
}

var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)
var _ signal.SignalWithNoOpCheck = (*Signal)(nil)
var _ signal.SignalWithPolicyEvaluation = (*Signal)(nil)
var _ signal.SignalWithAutoRetry = (*Signal)(nil)
var _ signal.SignalWithMaxRetries = (*Signal)(nil)
var _ signal.SignalWithMaxAutoRetries = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)
var _ signal.SignalWithSkipNoops = (*Signal)(nil)
var _ signal.SignalWithAutoApproveOnPoliciesPassing = (*Signal)(nil)
var _ signal.SignalWithOnRetry = (*Signal)(nil)
var _ signal.SignalWithApprovalValidation = (*Signal)(nil)

func (s *Signal) ValidateApproval(ctx workflow.Context) error {
	if s.ComponentID == "" || s.FlowID == "" {
		return nil
	}
	deploy, err := activities.AwaitGetInstallDeployForApplyStep(ctx, activities.GetInstallDeployForApplyStep{
		InstallWorkflowID: s.FlowID,
		ComponentID:       s.ComponentID,
	})
	if err != nil {
		return nil
	}
	superseded, err := activities.AwaitCheckDeploySuperseded(ctx, activities.CheckDeploySupersededRequest{
		DeployID: deploy.ID,
	})
	if err != nil {
		return nil
	}
	if superseded {
		return fmt.Errorf("a newer deploy for this component has completed since this plan was created")
	}
	return nil
}

func (s *Signal) OnRetry(ctx workflow.Context) error {
	deploy, err := activities.AwaitGetInstallDeployForApplyStep(ctx, activities.GetInstallDeployForApplyStep{
		InstallWorkflowID: s.FlowID,
		ComponentID:       s.ComponentID,
	})
	if err != nil {
		return nil
	}
	s.updateDeployStatusWithoutStatusSync(ctx, deploy.ID, app.InstallDeployStatusRetried, "deploy retried")
	return nil
}

func (s *Signal) IsNoOpCheckable() bool                 { return true }
func (s *Signal) RequiresPolicyEvaluation() bool        { return true }
func (s *Signal) AutoRetry() bool                       { return true }
func (s *Signal) MaxRetries() int                       { return 15 }
func (s *Signal) MaxAutoRetries(_ workflow.Context) int { return 3 }

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	if s.runnerJobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, s.runnerJobID)
		return nil
	}
	if jobID := activeCheckpointJobID(cancelCtx); jobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, jobID)
		var prep deployCheckpoint
		if ok, err := job.LoadCheckpoint(cancelCtx, checkpointDeploy, &prep); err == nil && ok && prep.DeployID != "" {
			s.updateDeployStatusWithoutStatusSync(cancelCtx, prep.DeployID, app.InstallDeployStatusCancelled, "teardown cancelled")
		}
	}
	return nil
}

func (s *Signal) SkipNoops(ctx workflow.Context) bool {
	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, s.InstallComponentID)
	if err != nil {
		return false
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return false
	}

	for _, ccc := range appCfg.ComponentConfigConnections {
		if ccc.ComponentID == s.ComponentID {
			return ccc.GetSkipNoops()
		}
	}
	return false
}

func (s *Signal) AutoApproveOnPoliciesPassing(ctx workflow.Context) bool {
	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, s.InstallComponentID)
	if err != nil {
		return false
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return false
	}

	for _, ccc := range appCfg.ComponentConfigConnections {
		if ccc.ComponentID == s.ComponentID {
			return ccc.GetAutoApproveOnPoliciesPassing()
		}
	}
	return false
}

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	installID := &s.InstallID
	if s.InstallID == "" {
		installID = nil
	}
	componentID := &s.ComponentID
	if s.ComponentID == "" {
		componentID = nil
	}
	return signal.SignalLifecycleContext{
		InstallID:    installID,
		ComponentID:  componentID,
		Operation:    "component-teardown",
		Stage:        "plan",
		WorkflowID:   s.LifecycleWorkflowID,
		WorkflowType: s.LifecycleWorkflowType,
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	// Validate install component exists
	_, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, s.InstallComponentID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, s.InstallComponentID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	if s.ComponentID == "" {
		return fmt.Errorf("component ID is required")
	}

	var prep deployCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointDeploy, &prep); err != nil {
		return errors.Wrap(err, "unable to load deploy checkpoint")
	}

	var installDeploy *app.InstallDeploy
	if prep.DeployID != "" {
		installDeploy, err = activities.AwaitGetDeployByDeployID(ctx, prep.DeployID)
		if err != nil {
			return errors.Wrap(err, "unable to get install deploy")
		}
	} else {
		latestDeploy, err := activities.AwaitGetLatestDeploy(ctx, activities.GetLatestDeployRequest{
			ComponentID:   s.ComponentID,
			InstallID:     install.ID,
			PreferApplied: true,
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}

			return fmt.Errorf("unable to get component build: %w", err)
		}

		build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, latestDeploy.ComponentBuildID)
		if err != nil {
			return errors.Wrap(err, "unable to get build")
		}

		installDeploy, err = activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
			InstallID:   install.ID,
			ComponentID: s.ComponentID,
			BuildID:     build.ID,
			Type:        app.InstallDeployTypeTeardown,
			WorkflowID:  s.FlowID,
			Role:        s.Role,
		})
		if err != nil {
			return fmt.Errorf("unable to create install deploy: %w", err)
		}

		prep.DeployID = installDeploy.ID
		if err := job.SaveCheckpoint(ctx, checkpointDeploy, prep); err != nil {
			return errors.Wrap(err, "unable to checkpoint deploy")
		}
	}

	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			s.updateDeployStatusWithoutStatusSync(updateCtx, installDeploy.ID, app.InstallDeployStatusCancelled, "teardown cancelled")
		}
	}()

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         s.WorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: "install_deploys",
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	var logStream *app.LogStream
	if prep.LogStreamID != "" {
		logStream, err = activities.AwaitGetLogStreamByLogStreamID(ctx, prep.LogStreamID)
		if err != nil {
			return errors.Wrap(err, "unable to get log stream")
		}
	} else {
		logStream, err = activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
			DeployID: installDeploy.ID,
		})
		if err != nil {
			return errors.Wrap(err, "unable to create log stream")
		}
		prep.LogStreamID = logStream.ID
		if err := job.SaveCheckpoint(ctx, checkpointDeploy, prep); err != nil {
			return errors.Wrap(err, "unable to checkpoint log stream")
		}
	}
	// NOTE: not closed so we can re-use this log stream for apply plan

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	installDeploy, err = activities.AwaitGetDeployByDeployID(ctx, installDeploy.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install deploy")
	}

	// Check if component type requires teardown
	_, err = activities.AwaitGetComponentByComponentID(ctx, installDeploy.InstallComponent.ComponentID)
	if err != nil {
		return errors.Wrap(err, "unable to get component")
	}

	// For components that don't require teardown, exit early
	// The original code checks if comp.Type is in an empty slice, which means all types require teardown
	shouldTeardown := true
	if !shouldTeardown {
		l.Info("nothing to teardown")
		return nil
	}

	l.Info("performing component teardown")
	if err := s.doTeardown(ctx, install, installDeploy, s.FlowStepID, s.SandboxMode); err != nil {
		if job.IsSuspended(err) {
			return err
		}
		return errors.Wrap(err, "unable to perform teardown")
	}

	return nil
}

func (s *Signal) doTeardown(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, flowStepID string, sandboxMode bool) error {
	err := s.pollForDeployableBuild(ctx, installDeploy.ID, installDeploy.ComponentBuildID)
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusNoop, "build is not deployable")
		return nil
	}

	if err := s.execSync(ctx, install, installDeploy); err != nil {
		if job.IsSuspended(err) {
			return err
		}
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "error syncing")
		return errors.Wrap(err, "unable to execute sync")
	}

	if err := s.execPlan(ctx, install, installDeploy, flowStepID, sandboxMode); err != nil {
		if job.IsSuspended(err) {
			return err
		}
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "error deploying")
		return errors.Wrap(err, "unable to execute deploy")
	}

	_ = activities.AwaitSetDeployPlannedAt(ctx, activities.SetDeployPlannedAtRequest{
		DeployID: installDeploy.ID,
	})

	s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusPendingApproval, "pending-approval")
	return nil
}

func (s *Signal) isBuildDeployable(bld *app.ComponentBuild) bool {
	return bld.Status == app.ComponentBuildStatusActive
}

func (s *Signal) pollForDeployableBuild(ctx workflow.Context, installDeployId, componentBuildID string) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	bld, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, componentBuildID)
	if err != nil {
		return errors.Wrap(err, "unable to get component build")
	}

	if s.isBuildDeployable(bld) {
		l.Info("build is deployable")
		return nil
	}

	l.Info("build is not yet deployable, polling")
	// check the build every 10 seconds for 1 hour
	sleepTimer := time.Second * 10
	maxAttempts := 360
	attempt := 0
	for {
		if attempt >= maxAttempts {
			return fmt.Errorf("build is not deployable after %d polling attempts", maxAttempts)
		}

		attempt++

		// Get the latest build
		bld, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, bld.ID)
		if err != nil {
			return fmt.Errorf("unable to get component build: %w", err)
		}

		// Check if the build is deployable
		if s.isBuildDeployable(bld) {
			return nil
		}

		if bld.Status == app.ComponentBuildStatusError {
			l.Error("component build is in an error state")
			return fmt.Errorf("component build is in an error state")
		}

		workflow.Sleep(ctx, sleepTimer)
	}
}

func (s *Signal) execSync(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy) error {
	var cp phaseCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointSync, &cp); err != nil {
		return errors.Wrap(err, "unable to load sync checkpoint")
	}
	if cp.Finalized {
		return nil
	}

	if err := imagesync.RunSyncJob(ctx, imagesync.RunSyncJobRequest{
		Install:       install,
		InstallDeploy: installDeploy,
		Status:        s.updateDeployStatusWithoutStatusSync,
		OnJobCreated:  func(jobID string) { s.runnerJobID = jobID },
		JobID:         cp.JobID,
		OnJobPrepared: func(ctx workflow.Context, jobID string) error {
			cp.JobID = jobID
			return job.SaveCheckpoint(ctx, checkpointSync, cp)
		},
	}); err != nil {
		return err
	}

	cp.Finalized = true
	if err := job.SaveCheckpoint(ctx, checkpointSync, cp); err != nil {
		return errors.Wrap(err, "unable to checkpoint sync finalization")
	}

	return nil
}

func (s *Signal) execPlan(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, stepID string, sandboxMode bool) error {
	var cp phaseCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointPlan, &cp); err != nil {
		return errors.Wrap(err, "unable to load plan checkpoint")
	}
	if cp.Finalized {
		return nil
	}

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	if cp.JobID == "" {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating deploy plan")
	}
	l.Info("planning and executing deploy")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		return fmt.Errorf("unable to get build: %w", err)
	}

	comp, err := activities.AwaitGetComponentByComponentID(ctx, build.ComponentID)
	if err != nil {
		return errors.Wrap(err, "unable to get component")
	}

	if cp.JobID == "" {
		prepared, err := s.preparePlanJob(ctx, install, installDeploy, build, sandboxMode)
		if err != nil {
			return err
		}
		cp = *prepared
		if err := job.SaveCheckpoint(ctx, checkpointPlan, cp); err != nil {
			return errors.Wrap(err, "unable to checkpoint plan job")
		}
	}
	s.runnerJobID = cp.JobID

	s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusExecuting, "creating plan")
	_, err = job.Execute(ctx, &job.ExecuteJobRequest{
		RunnerID: install.RunnerID,
		JobID:    cp.JobID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("queue-signal-%s-execute-job-%s", install.ID, cp.JobID),
	})
	if job.IsSuspended(err) {
		return err
	}
	if err != nil {
		msg := job.JobErrorMessage(err, "teardown plan job failed")
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, msg)
		l.Error("job did not succeed", zap.Error(err))
		return fmt.Errorf("unable to get install: %w", err)
	}

	planJob, err := activities.AwaitGetJobByID(ctx, cp.JobID)
	if err != nil {
		return errors.Wrap(err, "unable to get job")
	}

	var approvalTyp app.WorkflowStepApprovalType
	switch comp.Type {
	case app.ComponentTypeHelmChart:
		approvalTyp = app.HelmApprovalApprovalType
	case app.ComponentTypeTerraformModule:
		approvalTyp = app.TerraformPlanApprovalType
	case app.ComponentTypeKubernetesManifest:
		approvalTyp = app.KubernetesManifestApprovalType
	case app.ComponentTypePulumi:
		approvalTyp = app.PulumiApprovalType
	default:
		approvalTyp = app.NoopApprovalType
	}

	if planJob.Execution.Result == nil || (len(planJob.Execution.Result.Contents) < 1 && len(planJob.Execution.Result.ContentsGzip) < 1) {
		if planJob.Type != app.RunnerJobTypeJobNOOPDeploy {
			return errors.New("no plan returned from job")
		}
	}

	var contentsDisplay string
	if planJob.Type == app.RunnerJobTypeJobNOOPDeploy {
		l.Debug("job is a noop - no plan is expected")
		contentsDisplay = "{\"op\": \"noop\"}"
	} else {
		contentsDisplay, err = planJob.Execution.Result.GetContentsDisplayString()
		if err != nil {
			return errors.Wrap(err, "unable to get content display")
		}
	}

	if len(cp.Retain) > 0 {
		if err := activities.AwaitCheckTeardownPlan(ctx, &activities.CheckTeardownPlanRequest{
			JobID:  planJob.ID,
			Retain: cp.Retain,
		}); err != nil {
			s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, err.Error())
			return err
		}
	}

	if err := workflowstepapprovalrequest.Dispatch(ctx, &workflowstepapprovalrequest.Signal{
		InstallID:         install.ID,
		InstallWorkflowID: s.FlowID,
		WorkflowStepID:    stepID,
		OwnerID:           installDeploy.ID,
		OwnerType:         "install_deploys",
		RunnerJobID:       planJob.ID,
		ApprovalType:      approvalTyp,
		Plan:              contentsDisplay,
	}); err != nil {
		return errors.Wrap(err, "unable to create approval")
	}

	cp.Finalized = true
	if err := job.SaveCheckpoint(ctx, checkpointPlan, cp); err != nil {
		return errors.Wrap(err, "unable to checkpoint plan finalization")
	}

	return nil
}

func (s *Signal) preparePlanJob(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, build *app.ComponentBuild, sandboxMode bool) (*phaseCheckpoint, error) {
	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return nil, err
	}

	op := app.RunnerJobOperationTypeCreateTeardownPlan
	if installDeploy.Type == app.InstallDeployTypeApply {
		op = app.RunnerJobOperationTypeCreateApplyPlan
	}

	jobTyp := build.ComponentConfigConnection.Type.DeployPlanJobType()
	runnerJob, err := activities.AwaitCreateDeployJob(ctx, &activities.CreateDeployJobRequest{
		RunnerID:        install.RunnerGroup.Runners[0].ID,
		DeployID:        installDeploy.ID,
		Op:              op,
		Type:            jobTyp,
		LogStreamID:     logStreamID,
		TimeoutDuration: build.ComponentConfigConnection.GetDeployTimeout(),
		Metadata: map[string]string{
			"install_id":           install.ID,
			"deploy_id":            installDeploy.ID,
			"install_component_id": installDeploy.InstallComponentID,
			"component_id":         build.ComponentConfigConnection.ComponentID,
			"component_name":       build.ComponentConfigConnection.Component.Name,
			"operation":            string(op),
		},
	})
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create runner job")
		return nil, fmt.Errorf("unable to create runner job: %w", err)
	}
	s.runnerJobID = runnerJob.ID

	planCompositeErrorsEnabled := workflow.GetVersion(ctx, planCompositeErrorVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
	if planCompositeErrorsEnabled {
		_ = activities.AwaitSetInstallDeployPlanCompositeError(ctx, activities.SetInstallDeployPlanCompositeErrorRequest{
			InstallDeployID: installDeploy.ID,
		})
	}

	deployPlan, err := plan.AwaitCreateDeployPlan(ctx, &plan.CreateDeployPlanRequest{
		InstallDeployID: installDeploy.ID,
		InstallID:       install.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-create-deploy-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create deploy plan")
		if planCompositeErrorsEnabled && deployerrors.IsDeployPlanRenderFailed(err) {
			_ = activities.AwaitSetInstallDeployPlanCompositeError(ctx, activities.SetInstallDeployPlanCompositeErrorRequest{
				InstallDeployID: installDeploy.ID,
				ComponentName:   installDeploy.ComponentName,
				Stage:           deployerrors.PlanRenderStage(err),
				Detail:          deployerrors.PlanRenderDetail(err),
			})
		}
		return nil, errors.Wrap(err, "unable to create deploy plan")
	}

	planJSON, err := json.Marshal(deployPlan.Plan)
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create json from deploy plan")
		return nil, errors.Wrap(err, "unable to create json from plan")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			DeployPlan: deployPlan.Plan,
		},
	}); err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	if err := activities.AwaitRecordInstallRoleUsage(ctx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     install.ID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: deployPlan.RoleSelection,
	}); err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to record install role usage")
		return nil, fmt.Errorf("unable to record install role usage: %w", err)
	}

	return &phaseCheckpoint{
		JobID:  runnerJob.ID,
		Retain: teardownPlanRetain(deployPlan.Plan, build.ComponentConfigConnection.Type, sandboxMode),
	}, nil
}

func (s *Signal) updateDeployStatusWithoutStatusSync(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateDeployStatus(ctx, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            status,
		StatusDescription: message,
		SkipStatusSync:    true,
	}); err != nil {
		l.Error("unable to update deploy status",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateDeployStatusV2(ctx, statusactivities.UpdateDeployStatusV2Request{
		DeployID:          deployID,
		Status:            app.Status(status),
		StatusDescription: message,
		SkipStatusSync:    true,
	}); err != nil {
		l.Error("unable to update deploy status v2",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}
}
