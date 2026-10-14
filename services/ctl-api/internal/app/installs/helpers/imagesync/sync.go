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

type StatusFunc func(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string)

type RunSyncJobRequest struct {
	Install       *app.Install
	InstallDeploy *app.InstallDeploy
	Status        StatusFunc
	OnJobCreated  func(jobID string)
	// PlanWorkflowID and JobWorkflowID are part of in-flight Temporal histories
	// and must stay caller-supplied.
	PlanWorkflowID string
	JobWorkflowID  string
	JobID          string
	OnJobPrepared  func(ctx workflow.Context, jobID string) error
}

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
	jobID := req.JobID
	if jobID == "" {
		jobID, err = prepareSyncJob(ctx, req, status)
		if err != nil {
			return err
		}
	} else if req.OnJobCreated != nil {
		req.OnJobCreated(jobID)
	}

	execReq := &job.ExecuteJobRequest{
		RunnerID:   install.RunnerID,
		JobID:      jobID,
		WorkflowID: req.JobWorkflowID,
	}
	status(ctx, installDeploy.ID, app.InstallDeployStatusSyncing, "executing sync plan")
	if req.OnJobPrepared != nil {
		_, err = job.Execute(ctx, execReq, nil)
	} else {
		_, err = job.AwaitExecuteJob(ctx, execReq)
	}
	if job.IsSuspended(err) {
		return err
	}
	if err != nil {
		status(ctx, installDeploy.ID, app.InstallDeployStatusError, job.JobErrorMessage(err, "sync image job failed"))
		l.Error("error polling sync image job", zap.Error(err))
		return fmt.Errorf("unable to poll job: %w", err)
	}
	l.Info("sync image job was successfully completed")

	syncedJob, err := activities.AwaitGetJobByID(ctx, jobID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner job")
	}

	var ociArtOutputs state.OCIArtifactOutputs
	if err := mapstructure.Decode(syncedJob.ParsedOutputs["image"], &ociArtOutputs); err != nil {
		l.Error("error parsing oci artifact outputs", zap.Error(err))
		return errors.Wrap(err, "unable to parse oci artifact outputs")
	}

	if _, err := activities.AwaitCreateOCIArtifact(ctx, activities.CreateOCIArtifactRequest{
		OwnerID:       installDeploy.ID,
		OwnerType:     "install_deploys",
		Outputs:       ociArtOutputs,
		ReuseExisting: req.OnJobPrepared != nil && job.DemandEnabled(ctx),
	}); err != nil {
		return errors.Wrap(err, "unable to create oci artifact")
	}

	return nil
}

func prepareSyncJob(ctx workflow.Context, req RunSyncJobRequest, status StatusFunc) (string, error) {
	install, installDeploy := req.Install, req.InstallDeploy
	status(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating sync plan")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		status(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		return "", fmt.Errorf("unable to get build: %w", err)
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("unable to create runner job: %w", err)
	}
	if req.OnJobCreated != nil {
		req.OnJobCreated(runnerJob.ID)
	}

	runPlan, err := plan.AwaitCreateSyncPlan(ctx, &plan.CreateSyncPlanRequest{
		InstallID:       install.ID,
		InstallDeployID: installDeploy.ID,
		WorkflowID:      req.PlanWorkflowID,
	})
	if err != nil {
		status(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return "", errors.Wrap(err, "unable to create plan")
	}

	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		return "", errors.Wrap(err, "unable to create json")
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
		return "", fmt.Errorf("unable to save runner job plan: %w", err)
	}

	if req.OnJobPrepared != nil {
		if err := req.OnJobPrepared(ctx, runnerJob.ID); err != nil {
			return "", errors.Wrap(err, "unable to checkpoint sync job")
		}
	}

	return runnerJob.ID, nil
}

type SyncCheckpoint struct {
	DeployID    string `json:"deploy_id,omitempty"`
	LogStreamID string `json:"log_stream_id,omitempty"`
	JobID       string `json:"job_id,omitempty"`
	Finalized   bool   `json:"finalized,omitempty"`
}

type SyncRequest struct {
	Install           *app.Install
	ComponentID       string
	BuildID           string
	FlowID            string
	ParentLogStreamID string
	OnJobCreated      func(jobID string)
	WorkflowIDSuffix  string
	Checkpoint        *SyncCheckpoint
	SaveCheckpoint    func(ctx workflow.Context) error
}

func (r SyncRequest) save(ctx workflow.Context) error {
	if r.Checkpoint == nil || r.SaveCheckpoint == nil {
		return nil
	}
	return r.SaveCheckpoint(ctx)
}

func Sync(ctx workflow.Context, req SyncRequest) (deploy *app.InstallDeploy, err error) {
	cp := req.Checkpoint
	if cp == nil {
		cp = &SyncCheckpoint{}
	}

	var installDeploy *app.InstallDeploy
	if cp.DeployID != "" {
		installDeploy, err = activities.AwaitGetDeployByDeployID(ctx, cp.DeployID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install deploy")
		}
	} else {
		installDeploy, err = activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
			InstallID:   req.Install.ID,
			ComponentID: req.ComponentID,
			BuildID:     req.BuildID,
			Type:        app.InstallDeployTypeSync,
			WorkflowID:  req.FlowID,
			// Empty role: a non-empty one is a hard request with no fallback, and
			// an action's role is often not a valid deploy role for the image.
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to create install deploy")
		}
		cp.DeployID = installDeploy.ID
		if err := req.save(ctx); err != nil {
			return nil, errors.Wrap(err, "unable to checkpoint install deploy")
		}
	}
	if cp.Finalized {
		return installDeploy, nil
	}

	var logStream *app.LogStream
	if cp.LogStreamID != "" {
		logStream, err = activities.AwaitGetLogStreamByLogStreamID(ctx, cp.LogStreamID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get log stream")
		}
	} else {
		logStream, err = activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
			DeployID:          installDeploy.ID,
			ParentLogStreamID: req.ParentLogStreamID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to create log stream")
		}
		cp.LogStreamID = logStream.ID
		if err := req.save(ctx); err != nil {
			return nil, errors.Wrap(err, "unable to checkpoint log stream")
		}
	}
	defer func() {
		if !job.IsSuspended(err) {
			activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
		}
	}()

	syncCtx := cctx.SetLogStreamWorkflowContext(ctx, logStream)

	runReq := RunSyncJobRequest{
		Install:       req.Install,
		InstallDeploy: installDeploy,
		Status: func(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string) {
			updateDeployStatus(ctx, deployID, status, message, true)
		},
		OnJobCreated:   req.OnJobCreated,
		PlanWorkflowID: fmt.Sprintf("%s-create-oci-sync-plan%s", workflow.GetInfo(ctx).WorkflowExecution.ID, req.WorkflowIDSuffix),
		JobWorkflowID:  fmt.Sprintf("%s-execute-job%s", workflow.GetInfo(ctx).WorkflowExecution.ID, req.WorkflowIDSuffix),
		JobID:          cp.JobID,
	}
	if req.Checkpoint != nil && req.SaveCheckpoint != nil {
		runReq.OnJobPrepared = func(ctx workflow.Context, jobID string) error {
			cp.JobID = jobID
			return req.save(ctx)
		}
	}

	if err := RunSyncJob(syncCtx, runReq); err != nil {
		if job.IsSuspended(err) {
			return nil, err
		}
		updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to sync", false)
		return nil, err
	}

	updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusActive, "finished", false)

	cp.Finalized = true
	if err := req.save(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to checkpoint sync finalization")
	}

	return installDeploy, nil
}

func updateDeployStatus(
	ctx workflow.Context,
	deployID string,
	status app.InstallDeployStatus,
	message string,
	skipStatusSync bool,
) {
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
