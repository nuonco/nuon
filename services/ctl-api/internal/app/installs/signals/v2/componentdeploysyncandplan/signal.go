package componentdeploysyncandplan

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "component-deploy-sync-and-plan"

type Signal struct {
	InstallComponentID string
	InstallID          string
	DeployID           string
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

var (
	_ signal.SignalWithStepContext      = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithNoOpCheck        = (*Signal)(nil)
	_ signal.SignalWithPolicyEvaluation = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
	_ signal.SignalWithMaxRetries       = (*Signal)(nil)
	_ signal.SignalWithMaxAutoRetries   = (*Signal)(nil)
	_ signal.SignalWithCancel           = (*Signal)(nil)
)

func (s *Signal) IsNoOpCheckable() bool          { return true }
func (s *Signal) RequiresPolicyEvaluation() bool { return true }
func (s *Signal) AutoRetry() bool                { return true }
func (s *Signal) MaxRetries() int                { return 5 }

func (s *Signal) MaxAutoRetries(_ workflow.Context) int { return 3 }

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	if s.runnerJobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, s.runnerJobID)
	}
	return nil
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
		InstallID:   installID,
		ComponentID: componentID,
		Operation:   "component-deploy",
		Stage:       "plan",
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
		s.updateDeployStatusWithoutStatusSync(ctx, s.DeployID, app.InstallDeployStatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	var installDeploy *app.InstallDeploy
	if s.DeployID != "" {
		installDeploy, err = activities.AwaitGetDeployByDeployID(ctx, s.DeployID)
		if err != nil {
			return errors.Wrap(err, "unable to get deploy")
		}
		s.DeployID = installDeploy.ID
	} else {
		componentBuild, err := activities.AwaitGetComponentLatestBuildByComponentID(ctx, s.ComponentID)
		if err != nil {
			return fmt.Errorf("unable to get component build: %w", err)
		}

		installDeploy, err = activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
			InstallID:   install.ID,
			ComponentID: s.ComponentID,
			BuildID:     componentBuild.ID,
			Type:        app.InstallDeployTypeApply,
			WorkflowID:  s.FlowID,
			Role:        s.Role,
		})
		if err != nil {
			return fmt.Errorf("unable to create install deploy: %w", err)
		}
		s.DeployID = installDeploy.ID
	}

	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			s.updateDeployStatusWithoutStatusSync(updateCtx, installDeploy.ID, app.InstallDeployStatusCancelled, "deploy cancelled")
		}
	}()

	err = s.pollForDeployableBuild(ctx, installDeploy.ID, installDeploy.ComponentBuildID)
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusNoop, "build is not deployable")
		return errors.Wrap(err, "failed to poll for build")
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         s.WorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: "install_deploys",
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	defer func() {
		if pan := recover(); pan != nil {
			s.updateDeployStatusWithoutStatusSync(ctx, s.DeployID, app.InstallDeployStatusError, "internal error")
			panic(pan)
		}
	}()

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		DeployID: s.DeployID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	// NOTE: not closed so we can re-use this log stream for apply plan

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("syncing oci artifact")
	if err := s.execSync(ctx, install, installDeploy, s.SandboxMode); err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to sync")
		return errors.Wrap(err, "unable to execute sync")
	}

	l.Info("creating plan")
	if err := s.execPlan(ctx, install, installDeploy, s.FlowStepID, s.SandboxMode); err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to deploy")
		return errors.Wrap(err, "unable to execute deploy")
	}

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

func (s *Signal) execSync(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("syncing image into install OCI repository")
	s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating sync plan")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		return fmt.Errorf("unable to get build: %w", err)
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}

	runnerJob, err := activities.AwaitCreateSyncJob(ctx, &activities.CreateSyncJobRequest{
		DeployID:    installDeploy.ID,
		RunnerID:    install.RunnerID,
		Op:          app.RunnerJobOperationTypeExec,
		Type:        build.ComponentConfigConnection.Type.SyncJobType(),
		LogStreamID: logStreamID,
		Metadata: map[string]string{
			"install_id":           install.ID,
			"deploy_id":            installDeploy.ID,
			"install_component_id": installDeploy.InstallComponentID,
			"component_id":         build.ComponentConfigConnection.ComponentID,
			"component_name":       build.ComponentConfigConnection.Component.Name,
		},
	})
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}
	s.runnerJobID = runnerJob.ID

	// create the plan request
	runPlan, err := plan.AwaitCreateSyncPlan(ctx, &plan.CreateSyncPlanRequest{
		InstallID:       install.ID,
		InstallDeployID: installDeploy.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-create-oci-sync-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return errors.Wrap(err, "unable to create plan")
	}

	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	// Deprecated: for now we dual write both the plan json and the composite plan
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			SyncOCIPlan: runPlan,
		},
	}); err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	// queue job
	s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusSyncing, "executing sync plan")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID: install.RunnerID,
		JobID:    runnerJob.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-execute-job", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to poll job")
		l.Error("error polling sync image job", zap.Error(err))
		return fmt.Errorf("unable to poll job: %w", err)
	}
	l.Info("sync image job was successfully completed")

	// parse outputs and create OCI artifact record
	job, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner job")
	}

	var ociArtOutputs state.OCIArtifactOutputs
	if err := mapstructure.Decode(job.ParsedOutputs["image"], &ociArtOutputs); err != nil {
		l.Error("error parsing oci artifact outputs", zap.Error(err))
		return errors.Wrap(err, "unable to parse oci artifact outputs")
	}

	if _, err := activities.AwaitCreateOCIArtifact(ctx, activities.CreateOCIArtifactRequest{
		OwnerID:   installDeploy.ID,
		OwnerType: "install_deploys",
		Outputs:   ociArtOutputs,
	}); err != nil {
		return errors.Wrap(err, "unable to create oci artifact")
	}

	return nil
}

func (s *Signal) execPlan(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, stepID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating deploy plan")
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

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}

	op := app.RunnerJobOperationTypeCreateTeardownPlan
	if installDeploy.Type == app.InstallDeployTypeApply {
		op = app.RunnerJobOperationTypeCreateApplyPlan
	}

	jobTyp := build.ComponentConfigConnection.Type.DeployPlanJobType()
	runnerJob, err := activities.AwaitCreateDeployJob(ctx, &activities.CreateDeployJobRequest{
		RunnerID:    install.RunnerGroup.Runners[0].ID,
		DeployID:    installDeploy.ID,
		Op:          op,
		Type:        jobTyp,
		LogStreamID: logStreamID,
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
		return fmt.Errorf("unable to create runner job: %w", err)
	}
	s.runnerJobID = runnerJob.ID

	deployPlan, err := plan.AwaitCreateDeployPlan(ctx, &plan.CreateDeployPlanRequest{
		InstallDeployID: installDeploy.ID,
		InstallID:       install.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-create-deploy-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create deploy plan")
		return errors.Wrap(err, "unable to create deploy plan")
	}

	planJSON, err := json.Marshal(deployPlan.Plan)
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create json from deploy plan")
		return errors.Wrap(err, "unable to create json from plan")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			DeployPlan: deployPlan.Plan,
		},
	}); err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	if err := activities.AwaitRecordInstallRoleUsage(ctx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     install.ID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: deployPlan.RoleSelection,
	}); err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to record install role usage")
		return fmt.Errorf("unable to record install role usage: %w", err)
	}

	planJSON = nil

	s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusExecuting, "creating plan")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID: install.RunnerID,
		JobID:    runnerJob.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to execute runner job")
		l.Error("job did not succeed", zap.Error(err))
		return fmt.Errorf("unable to get install: %w", err)
	}

	job, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
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

	// all component types require a plan EXCEPT for docker builds
	if job.Execution.Result == nil || (len(job.Execution.Result.Contents) < 1 && len(job.Execution.Result.ContentsGzip) < 1) {
		if runnerJob.Type != app.RunnerJobTypeJobNOOPDeploy {
			return errors.New("no plan returned from job")
		}
	}

	var contentsDisplay string
	if runnerJob.Type == app.RunnerJobTypeJobNOOPDeploy {
		l.Debug("job is a noop - no plan is expected")
		contentsDisplay = "{\"op\": \"noop\"}"
	} else {
		contentsDisplay, err = job.Execution.Result.GetContentsDisplayString()
		if err != nil {
			return errors.Wrap(err, "unable to get content display")
		}
	}

	if _, err := activities.AwaitCreateStepApproval(ctx, &activities.CreateStepApprovalRequest{
		OwnerID:     installDeploy.ID,
		OwnerType:   "install_deploys",
		RunnerJobID: job.ID,
		StepID:      stepID,
		Plan:        contentsDisplay,
		Type:        approvalTyp,
	}); err != nil {
		return errors.Wrap(err, "unable to create approval")
	}

	return nil
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
