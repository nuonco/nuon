package componenthelmrecover

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
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

const SignalType signal.SignalType = "component-helm-release-recover"

// Signal recovers a Helm release that was left mid-operation. It is deliberately
// not a deploy: it applies no chart and changes no desired state, so it neither
// plans nor asks for approval. It exists to run one helm rollback (or, when
// nothing ever rolled out, one uninstall) on the runner and report what it did.
type Signal struct {
	signal.LifecycleBase

	InstallID             string
	InstallComponentID    string
	ComponentID           string
	InstallDeployID       string
	FlowID                string
	FlowStepID            string
	InstallWorkflowStepID string

	runnerJobID string
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.InstallWorkflowStepID = stepID
	s.FlowStepID = stepID
	s.FlowID = flowID
}

var (
	_ signal.SignalWithStepContext      = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithCancel           = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
	_ signal.SignalWithMaxRetries       = (*Signal)(nil)
)

// AutoRetry is off. A recovery either works or hit something a retry will not
// fix, and it is a break-glass operation an operator is already watching — retry
// loops on a cluster-mutating recovery are worse than an honest failure.
func (s *Signal) AutoRetry() bool { return false }

func (s *Signal) MaxRetries() int { return 1 }

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
		InstallID:    installID,
		ComponentID:  componentID,
		Operation:    "component-helm-recover",
		Stage:        "apply",
		WorkflowID:   s.LifecycleWorkflowID,
		WorkflowType: s.LifecycleWorkflowType,
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallDeployID == "" {
		return errors.New("install deploy id is required")
	}
	if _, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, s.InstallComponentID); err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, s.InstallComponentID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, s.InstallDeployID)
	if err != nil {
		return errors.Wrap(err, "unable to get install deploy")
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         s.InstallWorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: "install_deploys",
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	// The endpoint's deploy row has no log stream, and runner_jobs.log_stream_id
	// is a foreign key. Also mints the service account the runner writes with.
	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		DeployID: installDeploy.ID,
		StepID:   s.InstallWorkflowStepID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream for the recovery")
	}

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)

	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			s.updateDeployStatus(updateCtx, installDeploy.ID, app.InstallDeployStatusCancelled, "recovery cancelled")
		}
	}()

	if err := s.execRecover(ctx, install, installDeploy); err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to recover helm release")
		return errors.Wrap(err, "unable to recover helm release")
	}

	_ = activities.AwaitSetDeployAppliedAt(ctx, activities.SetDeployAppliedAtRequest{
		DeployID: installDeploy.ID,
	})

	// Not inactive: the dashboard reads the newest deploy of any type as the
	// component's state and treats inactive as torn down.
	s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusActive, "helm release recovered")

	return nil
}

func (s *Signal) execRecover(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "preparing helm release recovery")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		return fmt.Errorf("unable to get build: %w", err)
	}

	// Re-checked here: the endpoint's check can go stale before the step runs.
	if build.ComponentConfigConnection.Type != app.ComponentTypeHelmChart {
		return fmt.Errorf("component %s is a %s component, not a helm chart",
			build.ComponentConfigConnection.Component.Name, build.ComponentConfigConnection.Type)
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStreamID)
	}()

	op := app.RunnerJobOperationTypeApplyPlan
	jobTyp := app.RunnerJobTypeHelmChartDeploy

	runnerJob, err := activities.AwaitCreateDeployJob(ctx, &activities.CreateDeployJobRequest{
		RunnerID:        install.RunnerID,
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
			"recover_helm_release": "true",
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create runner job: %w", err)
	}
	s.runnerJobID = runnerJob.ID

	deployPlan, err := plan.AwaitCreateDeployPlan(ctx, &plan.CreateDeployPlanRequest{
		InstallDeployID: installDeploy.ID,
		InstallID:       install.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-create-recover-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create deploy plan")
	}

	if deployPlan.Plan == nil || deployPlan.Plan.HelmDeployPlan == nil {
		return errors.New("deploy plan has no helm section, so there is no release to recover")
	}

	// Stamped here rather than threaded through the builder: this is its only caller.
	deployPlan.Plan.HelmDeployPlan.RecoverRelease = true

	// A recovery reads the chart from the stored revision, so any apply contents
	// inherited from the plan builder would only mislead the runner.
	deployPlan.Plan.ApplyPlanContents = ""
	deployPlan.Plan.ApplyPlanDisplay = ""

	planJSON, err := json.Marshal(deployPlan.Plan)
	if err != nil {
		return errors.Wrap(err, "unable to create json from plan")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			DeployPlan: deployPlan.Plan,
		},
	}); err != nil {
		return fmt.Errorf("unable to store runner job plan: %w", err)
	}

	if err := activities.AwaitRecordInstallRoleUsage(ctx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     install.ID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: deployPlan.RoleSelection,
	}); err != nil {
		return fmt.Errorf("unable to record install role usage: %w", err)
	}

	planJSON = nil

	s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusExecuting, "recovering helm release")
	if _, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID: install.RunnerID,
		JobID:    runnerJob.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("queue-signal-%s-execute-job-%s", install.ID, runnerJob.ID),
	}); err != nil {
		l.Error("helm release recovery job did not succeed", zap.Error(err))
		return fmt.Errorf("%s", job.JobErrorMessage(err, "helm release recovery job failed"))
	}

	return nil
}

func (s *Signal) updateDeployStatus(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateDeployStatus(ctx, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            status,
		StatusDescription: message,
		SkipStatusSync:    false,
	}); err != nil {
		l.Error("unable to update deploy status",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateDeployStatusV2(ctx, statusactivities.UpdateDeployStatusV2Request{
		DeployID:          deployID,
		Status:            app.Status(status),
		StatusDescription: message,
		SkipStatusSync:    false,
	}); err != nil {
		l.Error("unable to update deploy status v2",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}
}
