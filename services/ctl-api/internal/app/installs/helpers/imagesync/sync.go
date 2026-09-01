package imagesync

import (
	"encoding/json"
	"fmt"

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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// StatusFunc records sync progress on the install deploy that owns it. Callers
// own it because they disagree on whether a status write should also drive a
// status sync.
type StatusFunc func(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string)

// RunSyncJobRequest is the input to RunSyncJob.
type RunSyncJobRequest struct {
	Install       *app.Install
	InstallDeploy *app.InstallDeploy

	// Status is optional; sync progress goes unrecorded without it.
	Status StatusFunc

	// OnJobCreated is called as soon as the sync job row exists and before it
	// is queued, so a caller can cancel that job when its workflow is
	// cancelled mid-sync.
	OnJobCreated func(jobID string)

	// PlanWorkflowID and JobWorkflowID name the child workflows this starts.
	// They are caller-supplied for two reasons: the existing callers' IDs are
	// part of in-flight workflow histories and must not change, and a caller
	// that syncs more than one component from a single workflow needs a
	// distinct pair per component.
	PlanWorkflowID string
	JobWorkflowID  string
}

// RunSyncJob copies the deploy's component build into the install's own OCI
// registry: it plans the copy, runs it as a runner job, and records the
// resulting artifact against the deploy.
//
// The deploy and the log stream on ctx must already exist. The caller owns them
// because a deploy driven by a workflow step also has to retarget that step at
// it, and because a caller running this inline nests the log stream under its
// own.
func RunSyncJob(ctx workflow.Context, req RunSyncJobRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	install, installDeploy := req.Install, req.InstallDeploy
	status := req.Status
	if status == nil {
		status = func(workflow.Context, string, app.InstallDeployStatus, string) {}
	}

	l.Info("syncing image into install OCI repository")
	status(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating sync plan")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		status(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
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
		status(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}
	if req.OnJobCreated != nil {
		req.OnJobCreated(runnerJob.ID)
	}

	// create the plan request
	runPlan, err := plan.AwaitCreateSyncPlan(ctx, &plan.CreateSyncPlanRequest{
		InstallID:       install.ID,
		InstallDeployID: installDeploy.ID,
		WorkflowID:      req.PlanWorkflowID,
	})
	if err != nil {
		status(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
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
		status(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return fmt.Errorf("unable to save runner job plan: %w", err)
	}

	// queue job
	status(ctx, installDeploy.ID, app.InstallDeployStatusSyncing, "executing sync plan")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   install.RunnerID,
		JobID:      runnerJob.ID,
		WorkflowID: req.JobWorkflowID,
	})
	if err != nil {
		status(ctx, installDeploy.ID, app.InstallDeployStatusError, job.JobErrorMessage(err, "sync image job failed"))
		l.Error("error polling sync image job", zap.Error(err))
		return fmt.Errorf("unable to poll job: %w", err)
	}
	l.Info("sync image job was successfully completed")

	// parse outputs
	syncedJob, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner job")
	}

	var ociArtOutputs state.OCIArtifactOutputs
	if err := mapstructure.Decode(syncedJob.ParsedOutputs["image"], &ociArtOutputs); err != nil {
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

// SyncRequest is the input to Sync.
type SyncRequest struct {
	Install     *app.Install
	ComponentID string
	BuildID     string

	// FlowID is the install workflow the sync belongs to, recorded on the
	// deploy. Empty for a sync that no install workflow drives.
	FlowID string

	// ParentLogStreamID nests the sync's logs under the caller's stream, so
	// they read as part of whatever triggered them.
	ParentLogStreamID string

	OnJobCreated func(jobID string)

	// WorkflowIDSuffix is appended to the child workflow IDs the sync starts,
	// to disambiguate them when one workflow syncs several components. Include
	// a leading separator.
	WorkflowIDSuffix string
}

// Sync creates a sync deploy for one image component and runs it now.
//
// It exists for callers with no workflow step of their own to hang a sync off —
// an image-backed action run, which can be triggered by a cron with no workflow
// at all — and produces the same install-visible record a deploy would: an
// install deploy of type sync, an OCI artifact, and logs nested under the
// caller's stream. Callers still own state generation afterwards, since only
// they know which state the sync has to be visible in.
func Sync(ctx workflow.Context, req SyncRequest) (*app.InstallDeploy, error) {
	installDeploy, err := activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
		InstallID:   req.Install.ID,
		ComponentID: req.ComponentID,
		BuildID:     req.BuildID,
		Type:        app.InstallDeployTypeSync,
		WorkflowID:  req.FlowID,
		// Role is deliberately left empty. A non-empty deploy role is treated
		// as a hard request by role selection with no fallback, so passing a
		// caller's own role (an action's, say) fails the sync outright when
		// that role is not a valid deploy role for the image component. An
		// empty one resolves the same way a deploy with no explicit role does.
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create install deploy")
	}

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		DeployID:          installDeploy.ID,
		ParentLogStreamID: req.ParentLogStreamID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()

	syncCtx := cctx.SetLogStreamWorkflowContext(ctx, logStream)

	if err := RunSyncJob(syncCtx, RunSyncJobRequest{
		Install:       req.Install,
		InstallDeploy: installDeploy,
		Status: func(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string) {
			updateDeployStatus(ctx, deployID, status, message, true)
		},
		OnJobCreated:   req.OnJobCreated,
		PlanWorkflowID: fmt.Sprintf("%s-create-oci-sync-plan%s", workflow.GetInfo(ctx).WorkflowExecution.ID, req.WorkflowIDSuffix),
		JobWorkflowID:  fmt.Sprintf("%s-execute-job%s", workflow.GetInfo(ctx).WorkflowExecution.ID, req.WorkflowIDSuffix),
	}); err != nil {
		updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to sync", false)
		return nil, err
	}

	updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusActive, "finished", false)

	return installDeploy, nil
}

func updateDeployStatus(
	ctx workflow.Context,
	deployID string,
	status app.InstallDeployStatus,
	message string,
	skipStatusSync bool,
) {
	// A status write is best-effort, so failures are logged rather than
	// returned — but the writes still happen when there is no logger.
	logFailure := func(msg string, err error) {
		l, lErr := log.WorkflowLogger(ctx)
		if lErr != nil {
			return
		}
		l.Error(msg, zap.String("deploy-id", deployID), zap.Error(err))
	}

	if err := activities.AwaitUpdateDeployStatus(ctx, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            status,
		StatusDescription: message,
		SkipStatusSync:    skipStatusSync,
	}); err != nil {
		logFailure("unable to update deploy status", err)
	}

	if err := statusactivities.AwaitUpdateDeployStatusV2(ctx, statusactivities.UpdateDeployStatusV2Request{
		DeployID:          deployID,
		Status:            app.Status(status),
		StatusDescription: message,
		SkipStatusSync:    skipStatusSync,
	}); err != nil {
		logFailure("unable to update deploy status v2", err)
	}
}
