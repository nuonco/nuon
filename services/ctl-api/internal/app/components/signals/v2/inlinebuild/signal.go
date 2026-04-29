package inlinebuild

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
	ComponentID string `json:"component_id" validate:"required"`
	BuildID     string `json:"build_id"` // optional; if set, skip build creation and trigger pre-created build

	// Optional VCS pinning: caller resolves these before creating the signal.
	GitRef                *string `json:"git_ref,omitempty"`
	VCSConnectionCommitID *string `json:"vcs_connection_commit_id,omitempty"`

	// Step context injected by the flow engine via SignalWithStepContext.
	FlowID string `json:"flow_id,omitempty"`
	StepID string `json:"step_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithAutoRetry = (*Signal)(nil)
var _ signal.SignalWithMaxAutoRetries = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType               { return SignalType }
func (s *Signal) AutoRetry() bool                       { return true }
func (s *Signal) MaxAutoRetries(_ workflow.Context) int { return 3 }

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Phase 1: Create build record (if not pre-created).
	buildID := s.BuildID
	if buildID == "" {
		cmp, err := activities.AwaitGetComponentByComponentID(ctx, s.ComponentID)
		if err != nil {
			return fmt.Errorf("unable to get component: %w", err)
		}

		build, err := activities.AwaitCreateComponentBuildRecord(ctx, activities.CreateComponentBuildRecordRequest{
			CreatedByID:           cmp.CreatedByID,
			ComponentID:           s.ComponentID,
			OrgID:                 cmp.OrgID,
			GitRef:                s.GitRef,
			VCSConnectionCommitID: s.VCSConnectionCommitID,
		})
		if err != nil {
			return fmt.Errorf("unable to queue component build: %w", err)
		}
		buildID = build.ID
	}

	// Phase 2: Execute build inline (no queue hop).
	return s.execBuild(ctx, buildID)
}

// execBuild runs the full build lifecycle inline.
func (s *Signal) execBuild(ctx workflow.Context, buildID string) error {
	s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, "creating build plan")

	logStream, err := activities.AwaitCreateLogStreamByBuildID(ctx, buildID)
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("executing inline build")
	currentApp, err := activities.AwaitGetComponentAppByComponentID(ctx, s.ComponentID)
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component app")
		return fmt.Errorf("unable to get component app: %w", err)
	}

	comp, err := activities.AwaitGetComponentByComponentID(ctx, s.ComponentID)
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component")
		return fmt.Errorf("unable to get component: %w", err)
	}

	build, err := activities.AwaitGetComponentBuildByID(ctx, buildID)
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component build")
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
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "component is not active")
		return notify(fmt.Errorf("component is not active"))
	}

	// Get full component with org/runner preloads.
	fullComp, err := activities.AwaitGetComponent(ctx, activities.GetComponentRequest{
		ComponentID: s.ComponentID,
	})
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component")
		return notify(fmt.Errorf("unable to get component: %w", err))
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}
	runnerJob, err := activities.AwaitCreateBuildJob(ctx, &activities.CreateBuildJobRequest{
		RunnerID:    fullComp.Org.RunnerGroup.Runners[0].ID,
		BuildID:     buildID,
		Op:          app.RunnerJobOperationTypeBuild,
		Type:        fullComp.Type.BuildJobType(),
		LogStreamID: logStreamID,
		Metadata: map[string]string{
			"component_id":       fullComp.ID,
			"component_build_id": buildID,
			"component_name":     fullComp.Name,
			"app_id":             currentApp.ID,
		},
	})
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to create job")
		return notify(fmt.Errorf("unable to create job: %w", err))
	}

	cloudProvider, _ := activities.AwaitGetCloudProvider(ctx, &activities.GetCloudProviderRequest{})
	runPlan, err := plan.AwaitCreateComponentBuildPlan(ctx, &plan.CreateComponentBuildPlanRequest{
		ComponentID:      fullComp.ID,
		ComponentBuildID: buildID,
		WorkflowID:       fmt.Sprintf("%s-create-build-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
		CloudProvider:    cloudProvider,
	})
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component build plan")
		return notify(errors.Wrap(err, "unable to create plan"))
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
			return notify(errors.Wrap(err, "unable to get GAR access token"))
		}
		runPlan.ContainerImagePullPlan.RepoCfg.OCIAuth = &configs.OCIRegistryAuth{
			Username: garToken.Username,
			Password: garToken.Password,
		}
	}

	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to convert build plan to JSON")
		return notify(fmt.Errorf("unable to convert plan to json: %w", err))
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			BuildPlan: runPlan,
		},
	}); err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to save job plan")
		return notify(fmt.Errorf("unable to save runner job plan: %w", err))
	}

	// Execute the build job.
	s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusBuilding, "building")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID: fullComp.Org.RunnerGroup.Runners[0].ID,
		JobID:    runnerJob.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", fullComp.ID, runnerJob.ID),
	})
	if err != nil {
		s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, fmt.Sprintf("build failed: %s", signal.HumanError(err)))
		return notify(fmt.Errorf("build job failed: %w", err))
	}

	s.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusActive, "build is active and ready to be deployed")
	return nil
}

// updateBuildStatus updates the build status.
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

// sendNotification sends email and slack notifications.
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
