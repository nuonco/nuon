package build

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type Signal struct {
	signal.Hooks
	ComponentID string `json:"component_id" validate:"required"`
	BuildID     string `json:"build_id" validate:"required"`
	SandboxMode bool   `json:"sandbox_mode"`
}

var _ signal.Signal = (*Signal)(nil)

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
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	s.Hooks.LogStreamID = logStream.ID
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("executing build")
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

	if comp.Status != app.ComponentStatusActive {
		s.updateBuildStatus(ctx, s.BuildID, app.ComponentBuildStatusError, "component is not active")
		return notify(fmt.Errorf("component is not active"))
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
		RunnerID:    comp.Org.RunnerGroup.Runners[0].ID,
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

	cloudProvider, _ := activities.AwaitGetCloudProvider(ctx, &activities.GetCloudProviderRequest{})
	runPlan, err := plan.AwaitCreateComponentBuildPlan(ctx, &plan.CreateComponentBuildPlanRequest{
		ComponentID:      comp.ID,
		ComponentBuildID: buildID,
		WorkflowID:       fmt.Sprintf("%s-create-build-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
		CloudProvider:    cloudProvider,
	})
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component build plan")
		return errors.Wrap(err, "unable to create plan")
	}

	if runPlan.ContainerImagePullPlan != nil &&
		runPlan.ContainerImagePullPlan.RepoCfg != nil &&
		runPlan.ContainerImagePullPlan.RepoCfg.RegistryType == configs.OCIRegistryTypeGAR {
		garToken, err := activities.AwaitGetGARAccessToken(ctx, &activities.GetGARAccessTokenRequest{
			ServiceAccountEmail:      runPlan.ContainerImagePullPlan.RepoCfg.ServiceAccountEmail,
			WorkloadIdentityProvider: runPlan.ContainerImagePullPlan.RepoCfg.WorkloadIdentityProvider,
		})
		if err != nil {
			s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get GAR access token")
			return errors.Wrap(err, "unable to get GAR access token")
		}
		runPlan.ContainerImagePullPlan.RepoCfg.OCIAuth = &configs.OCIRegistryAuth{
			Username: garToken.Username,
			Password: garToken.Password,
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
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID: comp.Org.RunnerGroup.Runners[0].ID,
		JobID:    runnerJob.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", comp.ID, runnerJob.ID),
	})
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "build did not complete successfully")
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

// sendNotification sends email and slack notifications
func (s *Signal) sendNotification(ctx workflow.Context, typ notifications.Type, appID string, vars map[string]string) {
	l := workflow.GetLogger(ctx)

	// Send email
	if err := sharedactivities.AwaitSendEmail(ctx, sharedactivities.SendNotificationRequest{
		AppID: appID,
		Type:  typ,
		Vars:  vars,
	}); err != nil {
		l.Error("unable to send email",
			zap.Error(err),
			zap.String("type", typ.String()))
	}

	// Send slack notification
	if err := sharedactivities.AwaitSendSlack(ctx, sharedactivities.SendNotificationRequest{
		AppID: appID,
		Type:  typ,
		Vars:  vars,
	}); err != nil {
		l.Error("unable to send slack notification",
			zap.Error(err),
			zap.String("type", typ.String()))
	}
}
