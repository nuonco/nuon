package build

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	createdsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/created"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/plan"
	orgprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/provision"
	orgreprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/reprovision"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/controlplanejob"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type Signal struct {
	ComponentID string `json:"component_id" validate:"required"`
	BuildID     string `json:"build_id" validate:"required"`
	SandboxMode bool   `json:"sandbox_mode"`

	runnerJobID string
}

var (
	_ signal.Signal           = (*Signal)(nil)
	_ signal.SignalWithCancel = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	if s.BuildID == "" {
		return errors.New("build_id is required")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusPlanning, "creating build plan")

	logStream, err := activities.AwaitCreateLogStreamByBuildID(ctx, s.BuildID)
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		closeCtx, cancel := workflow.NewDisconnectedContext(ctx)
		defer cancel()
		activities.AwaitCloseLogStreamByLogStreamID(closeCtx, logStream.ID)
	}()
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("executing build")
	preflightOptions := workflow.ActivityOptions{
		RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 1},
	}
	if _, err := activities.AwaitGetBuildGitSourceByBuildID(ctx, s.BuildID, &preflightOptions); err != nil {
		l.Error(err.Error())
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, err.Error())
		return err
	}

	currentApp, err := activities.AwaitGetComponentAppByComponentID(ctx, s.ComponentID)
	if err != nil {
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "unable to get component app")
		return fmt.Errorf("unable to get component app: %w", err)
	}

	comp, err := activities.AwaitGetComponentByComponentID(ctx, s.ComponentID)
	if err != nil {
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "unable to get component")
		return fmt.Errorf("unable to get component: %w", err)
	}

	// Ensure org is provisioned before building.
	if err := queueclient.EnsureQueueSignal(ctx, comp.OrgID, "orgs", orgprovision.SignalType, orgreprovision.SignalType); err != nil {
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "org provision not ready")
		return fmt.Errorf("org provision not ready: %w", err)
	}
	if comp.Status != app.ComponentStatusActive {
		createdSignal := &createdsignal.Signal{ComponentID: s.ComponentID}
		if err := createdSignal.Execute(ctx); err != nil {
			s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "unable to activate component")
			return fmt.Errorf("unable to activate component: %w", err)
		}
	}

	// A component is activated asynchronously by its `created` signal. During a
	// sync burst the build can be enqueued before that signal commits the active
	// status, so wait for the activation signal to finish before checking status
	// rather than racing it and hard-failing.
	if err := queueclient.EnsureQueueSignal(ctx, comp.ID, "components", createdsignal.SignalType); err != nil {
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "component activation not ready")
		return fmt.Errorf("component activation not ready: %w", err)
	}

	// Re-fetch to observe the status committed by the activation signal.
	comp, err = activities.AwaitGetComponentByComponentID(ctx, s.ComponentID)
	if err != nil {
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "unable to get component")
		return fmt.Errorf("unable to get component: %w", err)
	}

	build, err := activities.AwaitGetComponentBuildByID(ctx, s.BuildID)
	if err != nil {
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "unable to get component build")
		return fmt.Errorf("unable to get component build: %w", err)
	}

	notify := func(err error) error {
		s.sendNotification(ctx, notifications.NotificationsTypeComponentBuildFailed, currentApp.ID, map[string]string{
			"component_name": comp.Name,
			"app_name":       currentApp.Name,
			"created_by":     build.CreatedBy.Email,
		})
		return err
	}

	if err := s.execBuild(ctx, s.ComponentID, s.BuildID, currentApp, s.SandboxMode); err != nil {
		return notify(err)
	}

	return nil
}

// execBuild executes the component build workflow
func (s *Signal) execBuild(ctx workflow.Context, compID, buildID string, currentApp *app.App, sandboxMode bool) error {
	comp, err := activities.AwaitGetComponent(ctx, activities.GetComponentRequest{
		ComponentID: compID,
	})
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component")
		return fmt.Errorf("unable to get component: %w", err)
	}

	// create the job
	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}
	runnerJob, err := activities.AwaitCreateBuildJob(ctx, &activities.CreateBuildJobRequest{
		BuildID:     buildID,
		Op:          app.RunnerJobOperationTypeBuild,
		Type:        comp.Type.BuildJobType(),
		LogStreamID: logStreamID,
		Metadata: map[string]string{
			"component_id":       comp.ID,
			"component_build_id": buildID,
			"component_name":     comp.Name,
			"app_id":             currentApp.ID,
		},
	})
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to create job")
		return fmt.Errorf("unable to create job: %w", err)
	}
	s.runnerJobID = runnerJob.ID
	if runnerJob.RunnerID == "" {
		if runnerJob.Executor != app.RunnerJobExecutorControlPlane {
			s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "no runners available in runner group")
			_ = activities.AwaitUpdateJobStatus(ctx, &activities.UpdateJobStatusRequest{
				JobID:             runnerJob.ID,
				Status:            app.RunnerJobStatusFailed,
				StatusDescription: "no runners available in runner group",
			})
			return fmt.Errorf("no runners available in runner group for org %s", comp.Org.ID)
		}
	}

	cloudProvider, _ := activities.AwaitGetCloudProvider(ctx, &activities.GetCloudProviderRequest{})
	managementIAMRoleARN, _ := activities.AwaitGetManagementIAMRoleARN(ctx, &activities.GetManagementIAMRoleARNRequest{})
	runPlan, err := plan.AwaitCreateComponentBuildPlan(ctx, &plan.CreateComponentBuildPlanRequest{
		ComponentID:          comp.ID,
		ComponentBuildID:     buildID,
		WorkflowID:           fmt.Sprintf("%s-create-build-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
		CloudProvider:        cloudProvider,
		ManagementIAMRoleARN: managementIAMRoleARN,
		IsControlPlaneBuild:  runnerJob.Executor == app.RunnerJobExecutorControlPlane,
	})
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component build plan")
		return errors.Wrap(err, "unable to create plan")
	}

	if runPlan.ContainerImagePullPlan != nil {
		if err := sharedactivities.EnsureGARAuth(ctx, runPlan.ContainerImagePullPlan.RepoCfg); err != nil {
			s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get GAR access token")
			return errors.Wrap(err, "unable to get GAR access token")
		}
	}

	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get convert build plan to JSON")
		return fmt.Errorf("unable to convert plan to json: %w", err)
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			BuildPlan: runPlan,
		},
	}); err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to save job plan")
		return fmt.Errorf("unable to save runner job plan: %w", err)
	}

	// wait for the job
	s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusBuilding, "building")
	if runnerJob.Executor == app.RunnerJobExecutorControlPlane {
		err = controlplanejob.AwaitExecuteControlPlaneJob(ctx, &controlplanejob.ExecuteRequest{JobID: runnerJob.ID}, &workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("control-plane-%s-execute-job-%s", comp.ID, runnerJob.ID),
		})
	} else {
		_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
			RunnerID: runnerJob.RunnerID,
			JobID:    runnerJob.ID,
		}, &workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("queue-signal-%s-execute-job-%s", comp.ID, runnerJob.ID),
		})
	}
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, fmt.Sprintf("build failed: %s", signal.HumanError(err)))
		return fmt.Errorf("build job failed: %w", err)
	}

	s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusActive, "build is active and ready to be deployed")
	return nil
}

// updateBuildStatus updates the build status
func (s *Signal) updateBuildStatus(ctx workflow.Context, bldID string, status app.ComponentBuildStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)
	err := activities.AwaitUpdateBuildStatus(ctx, activities.UpdateBuildStatus{
		BuildID:           bldID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err != nil {
		l.Error("unable to update build status",
			zap.String("build-id", bldID),
			zap.Error(err))
		return
	}

	err = statusactivities.AwaitUpdateBuildStatusV2(ctx, statusactivities.UpdateBuildStatusV2{
		BuildID:           bldID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err != nil {
		l.Error("unable to update build status v2",
			zap.String("build-id", bldID),
			zap.Error(err))
		return
	}
}

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	s.updateBuildStatus(cancelCtx, s.BuildID, app.ComponentBuildStatus(app.StatusCancelled), "build cancelled")
	if s.runnerJobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, s.runnerJobID)
	}
	return nil
}

func (s *Signal) sendNotification(ctx workflow.Context, typ notifications.Type, appID string, vars map[string]string) {
	l := workflow.GetLogger(ctx)

	if err := sharedactivities.AwaitSendEmail(ctx, sharedactivities.SendNotificationRequest{
		AppID: appID,
		Type:  typ,
		Vars:  vars,
	}); err != nil {
		l.Error("unable to send email",
			zap.Error(err),
			zap.String("type", typ.String()))
	}
}
