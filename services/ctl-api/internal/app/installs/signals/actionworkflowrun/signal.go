package actionworkflowrun

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/distribution/reference"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/stategen"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "install-action-workflow-run"
const runbookEventOutputsVersion = "runbook-event-outputs-v1"
const preparationCompositeErrorVersion = "action-preparation-composite-error-v1"
const actionImageDepSyncVersion = "action-image-dep-sync-v1"

type Signal struct {
	signal.LifecycleBase

	InstallID               string
	InstallWorkflowID       string
	WorkflowStepID          string
	InstallActionWorkflowID string
	AdhocActionRunID        string
	TriggerType             app.ActionWorkflowTriggerType
	TriggeredByID           string
	TriggeredByType         string
	RunEnvVars              map[string]string
	Role                    string

	runnerJobID string
}

var (
	_ signal.Signal                     = &Signal{}
	_ signal.SignalWithStepContext      = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithCancel           = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
	_ signal.SignalWithMaxRetries       = (*Signal)(nil)
	_ signal.SignalWithMaxAutoRetries   = (*Signal)(nil)
	_ signal.SignalWithOnRetry          = (*Signal)(nil)
)

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	if s.runnerJobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, s.runnerJobID)
		return nil
	}
	if jobID, runID := activeCheckpoint(cancelCtx); jobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, jobID)
		if runID != "" {
			s.updateActionRunStatus(cancelCtx, runID, app.InstallActionRunStatusCancelled, "action workflow run cancelled")
		}
	}
	return nil
}

func (s *Signal) OnRetry(ctx workflow.Context) error {
	// For adhoc runs, mark the existing run as retried.
	if s.AdhocActionRunID != "" {
		s.updateActionRunStatus(ctx, s.AdhocActionRunID, app.InstallActionRunStatusRetried, "retrying")
	}
	// Regular runs create the run during Execute — the old run was already
	// marked as error and a new run will be created on the retry clone.
	return nil
}

// AutoRetry enables the retry path in handleStepError so that failed action
// steps land at StepAwaitRetry instead of StepStop.
func (s *Signal) AutoRetry() bool { return true }

// MaxRetries is the total retry budget (auto + manual).
func (s *Signal) MaxRetries() int { return 3 }

// MaxAutoRetries returns 0 so auto-retries are immediately exhausted and the
// step goes straight to "awaiting retry or skip" for user action.
func (s *Signal) MaxAutoRetries(_ workflow.Context) int { return 0 }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID:    &s.InstallID,
		Operation:    "action-workflow-run",
		WorkflowID:   s.LifecycleWorkflowID,
		WorkflowType: s.LifecycleWorkflowType,
	}
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID = stepID
	s.InstallWorkflowID = flowID
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.AdhocActionRunID != "" {
		_, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, s.AdhocActionRunID)
		if err != nil {
			return fmt.Errorf("unable to get adhoc action run: %w", err)
		}
		return nil
	}

	if s.InstallActionWorkflowID == "" {
		return fmt.Errorf("install action workflow id is required")
	}

	// Validate install action workflow exists
	_, err := activities.AwaitGetInstallActionWorkflowByID(ctx, s.InstallActionWorkflowID)
	if err != nil {
		return fmt.Errorf("unable to get install action workflow: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	if s.AdhocActionRunID != "" {
		return s.executeAdhocRun(ctx)
	}

	l.Info("executing action workflow run signal",
		zap.String("install_action_workflow_id", s.InstallActionWorkflowID))

	installActionWorkflow, err := activities.AwaitGetInstallActionWorkflowByID(ctx, s.InstallActionWorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get install action workflow")
	}

	slimInstall, err := activities.AwaitGetSlimInstallByInstallID(ctx, installActionWorkflow.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	var prep runCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointRun, &prep); err != nil {
		return errors.Wrap(err, "unable to load action run checkpoint")
	}

	actionWorkflowRunID := prep.RunID
	if actionWorkflowRunID == "" {
		found, err := activities.AwaitActionWorkflowInAppConfig(ctx, activities.ActionWorkflowInAppConfigRequest{
			AppConfigID:      slimInstall.AppConfigID,
			ActionWorkflowID: installActionWorkflow.ActionWorkflowID,
		})
		if err != nil {
			return errors.Wrap(err, "unable to check action workflow in app config")
		}

		if !found {
			return fmt.Errorf("action workflow %s is not part of install's current app config", installActionWorkflow.ActionWorkflowID)
		}

		actionWorkflowRun, err := activities.AwaitCreateActionWorkflowRun(ctx, &activities.CreateActionWorkflowRunRequest{
			InstallActionWorkflowID: installActionWorkflow.ID,
			ActionWorkflowID:        installActionWorkflow.ActionWorkflowID,
			InstallID:               installActionWorkflow.InstallID,
			InstallWorkflowID:       s.InstallWorkflowID,
			TriggerType:             s.TriggerType,
			TriggeredByID:           s.TriggeredByID,
			TriggeredByType:         s.TriggeredByType,
			RunEnvVars:              generics.ToPtrStringMap(s.RunEnvVars),
			Role:                    s.Role,
		})
		if err != nil {
			return errors.Wrap(err, "unable to create action workflow run")
		}
		actionWorkflowRunID = actionWorkflowRun.ID
		prep.RunID = actionWorkflowRunID
		if err := job.SaveCheckpoint(ctx, checkpointRun, prep); err != nil {
			return errors.Wrap(err, "unable to checkpoint action workflow run")
		}
		if s.TriggeredByType == "runbook" && workflow.GetVersion(ctx, runbookEventOutputsVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
			if err := activities.AwaitRenderRunbookActionEventOutputs(ctx, activities.RenderRunbookActionEventOutputsRequest{ActionWorkflowRunID: actionWorkflowRunID}); err != nil {
				return errors.Wrap(err, "unable to render runbook event outputs")
			}
		}
	}

	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			s.updateActionRunStatus(updateCtx, actionWorkflowRunID, app.InstallActionRunStatusCancelled, "action workflow run cancelled")
		}
	}()

	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   actionWorkflowRunID,
			StepTargetType: "install_action_workflow_runs", // plugins.TableName would require db instance
		}); err != nil {
			return errors.Wrap(err, "unable to update install action workflow")
		}
	}

	return errors.Wrap(s.executeActionWorkflowRun(ctx, slimInstall.ID, slimInstall.Metadata, actionWorkflowRunID), "unable to execute action workflow run")
}

func (s *Signal) executeAdhocRun(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)
	l.Info("executing adhoc action workflow run signal",
		zap.String("adhoc_action_run_id", s.AdhocActionRunID))

	var prep runCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointRun, &prep); err != nil {
		return errors.Wrap(err, "unable to load action run checkpoint")
	}

	if prep.RunID == "" {
		if s.TriggeredByType == "runbook" && workflow.GetVersion(ctx, runbookEventOutputsVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
			if err := activities.AwaitRenderRunbookActionEventOutputs(ctx, activities.RenderRunbookActionEventOutputsRequest{ActionWorkflowRunID: s.AdhocActionRunID}); err != nil {
				return errors.Wrap(err, "unable to render runbook event outputs")
			}
		}
		prep.RunID = s.AdhocActionRunID
		if err := job.SaveCheckpoint(ctx, checkpointRun, prep); err != nil {
			return errors.Wrap(err, "unable to checkpoint action workflow run")
		}
	}

	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, s.AdhocActionRunID)
	if err != nil {
		return errors.Wrap(err, "unable to get adhoc action run")
	}

	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			s.updateActionRunStatus(updateCtx, run.ID, app.InstallActionRunStatusCancelled, "adhoc action workflow run cancelled")
		}
	}()

	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   run.ID,
			StepTargetType: "install_action_workflow_runs",
		}); err != nil {
			return errors.Wrap(err, "unable to update install action workflow")
		}
	}

	install, err := activities.AwaitGetByInstallID(ctx, run.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install for adhoc run")
	}

	return s.executeActionWorkflowRun(ctx, install.ID, install.Metadata, run.ID)
}

func (s *Signal) executeActionWorkflowRun(ctx workflow.Context, installID string, metadata pgtype.Hstore, actionWorkflowRunID string) error {
	l := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, actionWorkflowRunID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}
	preparationCompositeErrorsEnabled := workflow.GetVersion(ctx, preparationCompositeErrorVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
	if preparationCompositeErrorsEnabled {
		if err := activities.AwaitSetActionWorkflowRunPreparationCompositeError(ctx, activities.SetActionWorkflowRunPreparationCompositeErrorRequest{
			RunID: run.ID,
		}); err != nil {
			l.Warn("unable to clear action workflow run preparation composite error", zap.Error(err))
		}
	}

	var prep runCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointRun, &prep); err != nil {
		return errors.Wrap(err, "unable to load action run checkpoint")
	}

	parentLS, _ := cctx.GetLogStreamWorkflow(ctx)

	lsReq := activities.CreateLogStreamRequest{
		ActionWorkflowRunID: actionWorkflowRunID,
	}
	if parentLS != nil {
		lsReq.ParentLogStreamID = parentLS.ID
	}

	s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusInProgress, "in-progress")

	var ls *app.LogStream
	if prep.LogStreamID != "" {
		ls, err = activities.AwaitGetLogStreamByLogStreamID(ctx, prep.LogStreamID)
		if err != nil {
			return errors.Wrap(err, "unable to get log stream")
		}
	} else {
		ls, err = activities.AwaitCreateLogStream(ctx, lsReq)
		if err != nil {
			return errors.Wrap(err, "unable to create log stream")
		}
		prep.RunID = actionWorkflowRunID
		prep.LogStreamID = ls.ID
		if err := job.SaveCheckpoint(ctx, checkpointRun, prep); err != nil {
			return errors.Wrap(err, "unable to checkpoint log stream")
		}
	}

	if job.DemandEnabled(ctx) {
		return s.runActionSequence(cctx.SetLogStreamWorkflowContext(ctx, ls), installID, run, ls.ID, preparationCompositeErrorsEnabled)
	}
	defer activities.AwaitCloseLogStreamByLogStreamID(ctx, ls.ID)

	ctx = cctx.SetLogStreamWorkflowContext(ctx, ls)

	l, err = workflow.GetLogger(ctx), nil
	if err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create log stream")
		return errors.Wrap(err, "unable to set log stream on context")
	}

	ls, err = cctx.GetLogStreamWorkflow(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get log stream")
	}

	if workflow.GetVersion(ctx, actionImageDepSyncVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
		if err := s.syncActionImageDeps(ctx, run, ls.ID); err != nil {
			if preparationCompositeErrorsEnabled {
				s.recordPreparationCompositeError(ctx, run.ID, err)
			}
			s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to sync action image dependencies")
			return errors.Wrap(err, "unable to sync action image dependencies")
		}
	}

	jobID, err := s.prepareActionJob(ctx, installID, run, ls.ID, preparationCompositeErrorsEnabled)
	if err != nil {
		return err
	}

	l.Info("executing runner job")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID: run.Install.RunnerID,
		JobID:    jobID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: "actions-install-run-exec-job" + run.ID,
	})
	if err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, job.JobErrorMessage(err, "action workflow job failed"))
		return errors.Wrap(err, "runner job failed")
	}

	return s.finalizeActionRun(ctx, installID, run.ID)
}

func (s *Signal) finalizeActionRun(ctx workflow.Context, installID, runID string) error {
	s.updateActionRunStatus(ctx, runID, app.InstallActionRunStatusFinished, "finished")

	if s.InstallActionWorkflowID != "" {
		if err := stategen.HintOrGenerate(ctx, stategen.Request{
			InstallID:       installID,
			Targets:         statemanager.TargetsForHint(statemanager.HintActionRan, s.InstallActionWorkflowID),
			ForceAll:        true,
			TriggeredByID:   runID,
			TriggeredByType: "install_action_workflow_runs",
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Signal) prepareActionJob(ctx workflow.Context, installID string, run *app.InstallActionWorkflowRun, logStreamID string, preparationCompositeErrorsEnabled bool) (string, error) {
	planResponse, err := s.createActionPlan(ctx, run, preparationCompositeErrorsEnabled)
	if err != nil {
		return "", err
	}

	// image-backed actions: verify the org and runner can host the container
	// before dispatching. A plan that already carries a digest resolved the
	// image inside the install's own registry, so there is nothing to mirror.
	if planResponse.Plan.SourceImage != "" {
		if err := s.checkImageActionSupported(ctx, run); err != nil {
			s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, err.Error())
			return "", errors.Wrap(err, "unable to prepare image-backed action")
		}

		if planResponse.Plan.ImageDigestRef == "" {
			if err := s.mirrorActionImage(ctx, run, logStreamID, planResponse.Plan); err != nil {
				s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, err.Error())
				return "", errors.Wrap(err, "unable to prepare image-backed action")
			}
		}
	}

	return s.createActionJob(ctx, installID, run, logStreamID, planResponse, preparationCompositeErrorsEnabled)
}

func (s *Signal) createActionPlan(ctx workflow.Context, run *app.InstallActionWorkflowRun, preparationCompositeErrorsEnabled bool) (*plan.CreateActionPlanResponse, error) {
	workflow.GetLogger(ctx).Info("creating plan for executing action run")
	created, err := plan.AwaitCreateActionWorkflowRunPlan(ctx, &plan.CreateActionRunPlanRequest{
		ActionWorkflowRunID: run.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-create-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		if preparationCompositeErrorsEnabled {
			s.recordPreparationCompositeError(ctx, run.ID, err)
		}
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create plan")
		return nil, errors.Wrap(err, "unable to create plan")
	}
	return &created, nil
}

func (s *Signal) createActionJob(ctx workflow.Context, installID string, run *app.InstallActionWorkflowRun, logStreamID string, planResponse *plan.CreateActionPlanResponse, preparationCompositeErrorsEnabled bool) (string, error) {
	l := workflow.GetLogger(ctx)

	// execute job
	l.Info("creating runner job to execute action")
	runnerJob, err := activities.AwaitCreateActionWorkflowRunRunnerJob(ctx, &activities.CreateActionWorkflowRunRunnerJob{
		ActionWorkflowRunID: run.ID,
		RunnerID:            run.Install.RunnerID,
		LogStreamID:         logStreamID,
		Metadata: map[string]string{
			"install_id":             installID,
			"action_workflow_name":   run.ActionWorkflowConfig.ActionWorkflow.Name,
			"action_workflow_run_id": run.ID,
			"action_workflow_id":     run.ActionWorkflowConfig.ActionWorkflowID,
		},
	})
	if err != nil {
		if preparationCompositeErrorsEnabled {
			s.recordPreparationCompositeError(ctx, run.ID, err)
		}
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return "", errors.Wrap(err, "unable to create runner job")
	}
	s.runnerJobID = runnerJob.ID

	// save runner job plan
	planJSON, err := json.Marshal(planResponse.Plan)
	if err != nil {
		if preparationCompositeErrorsEnabled {
			s.recordPreparationCompositeError(ctx, run.ID, err)
		}
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return "", errors.Wrap(err, "unable to convert plan to json")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: plantypes.CompositePlan{ActionWorkflowRunPlan: planResponse.Plan},
	}); err != nil {
		if preparationCompositeErrorsEnabled {
			s.recordPreparationCompositeError(ctx, run.ID, err)
		}
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to save job plan")
		return "", errors.Wrap(err, "unable to save runner job plan")
	}

	if err := activities.AwaitRecordInstallRoleUsage(ctx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     installID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: planResponse.RoleSelection,
	}); err != nil {
		if preparationCompositeErrorsEnabled {
			s.recordPreparationCompositeError(ctx, run.ID, err)
		}
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to record install role usage")
		return "", errors.Wrap(err, "unable to record install role usage")
	}

	return runnerJob.ID, nil
}

func (s *Signal) recordPreparationCompositeError(ctx workflow.Context, runID string, runErr error) {
	if err := activities.AwaitSetActionWorkflowRunPreparationCompositeError(ctx, activities.SetActionWorkflowRunPreparationCompositeErrorRequest{
		RunID:  runID,
		Detail: signal.HumanError(runErr),
	}); err != nil {
		workflow.GetLogger(ctx).Warn("unable to record action workflow run preparation composite error", zap.Error(err))
	}
}

// checkImageActionSupported gates image-backed actions on the install's
// runner platform. It runs for every image-backed action, including ones
// that skip mirroring, so neither path can bypass the gate.
func (s *Signal) checkImageActionSupported(ctx workflow.Context, run *app.InstallActionWorkflowRun) error {
	platform := run.Install.RunnerGroup.Platform
	if !supportedImageActionPlatform(platform) {
		return fmt.Errorf("image-backed actions are only supported on AWS, Azure, and GCP VM runners; runner platform %q is not supported", platform)
	}

	return nil
}

// mirrorActionImage mirrors an image-backed action's app-authored image into
// the install registry via an oci-sync job. It pins the plan to the mirrored
// digest; any error here prevents the action job from being dispatched at all.
func (s *Signal) mirrorActionImage(ctx workflow.Context, run *app.InstallActionWorkflowRun, logStreamID string, awPlan *plantypes.ActionWorkflowRunPlan) error {
	jobID, err := s.prepareMirrorJob(ctx, run, logStreamID, awPlan)
	if err != nil {
		return err
	}

	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   run.Install.RunnerID,
		JobID:      jobID,
		WorkflowID: "action-image-sync-exec-job-" + run.ID,
	}, nil)
	if err != nil {
		return errors.Wrap(err, "image sync job failed")
	}

	digestRef, err := s.finalizeMirrorJob(ctx, run, jobID)
	if err != nil {
		return err
	}
	awPlan.ImageDigestRef = digestRef

	return nil
}

func (s *Signal) prepareMirrorJob(ctx workflow.Context, run *app.InstallActionWorkflowRun, logStreamID string, awPlan *plantypes.ActionWorkflowRunPlan) (string, error) {
	l := workflow.GetLogger(ctx)

	src, srcTag, err := parseActionImageSource(awPlan.SourceImage)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse action image reference")
	}

	l.Info("mirroring image-backed action image into install registry",
		zap.String("source_image", awPlan.SourceImage))

	syncJob, err := activities.AwaitCreateActionImageSyncJob(ctx, &activities.CreateActionImageSyncJobRequest{
		ActionWorkflowRunID: run.ID,
		RunnerID:            run.Install.RunnerID,
		LogStreamID:         logStreamID,
		Metadata: map[string]string{
			"install_id":             run.InstallID,
			"action_workflow_run_id": run.ID,
			"source_image":           awPlan.SourceImage,
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to create image sync job")
	}

	syncPlan := &plantypes.SyncOCIPlan{
		Src:    src,
		SrcTag: srcTag,
		Dst:    awPlan.ImageRegistry,
		DstTag: awPlan.ImageTag,
	}
	// A sandboxed sync job reports whatever the plan carries, so it has to carry
	// a digest-pinned ref: resolveMirroredDigestRef fails closed on a missing or
	// unpinned one, which would make every image-backed action fail in sandbox
	// mode rather than exercise the path.
	if awPlan.SandboxMode != nil {
		syncPlan.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: plantypes.FakeOCISyncOutputs(
				strings.TrimSuffix(awPlan.ImageRegistry.LoginServer+"/"+awPlan.ImageRegistry.Repository, "/"),
				awPlan.ImageTag,
			),
		}
	}

	syncPlanJSON, err := json.Marshal(syncPlan)
	if err != nil {
		return "", errors.Wrap(err, "unable to marshal sync plan")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         syncJob.ID,
		PlanJSON:      string(syncPlanJSON),
		CompositePlan: plantypes.CompositePlan{SyncOCIPlan: syncPlan},
	}); err != nil {
		return "", errors.Wrap(err, "unable to save sync job plan")
	}

	return syncJob.ID, nil
}

func (s *Signal) finalizeMirrorJob(ctx workflow.Context, run *app.InstallActionWorkflowRun, jobID string) (string, error) {
	l := workflow.GetLogger(ctx)

	// Bind execution to the exact manifest just mirrored: read the resolved
	// digest-pinned ref from the sync job outputs so the runner pulls by digest
	// rather than the mutable tag. This fails closed. If the digest can't be
	// resolved and validated we don't dispatch the action job at all, otherwise
	// the runner would execute whatever the mutable tag points at.
	digestRef, outputs, err := resolveMirroredDigestRef(ctx, jobID)
	if err != nil {
		return "", err
	}

	if workflow.GetVersion(ctx, actionImageDepSyncVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
		// One attempt: retried runs that re-mirror hit a unique owner key.
		if _, err := activities.AwaitCreateOCIArtifact(ctx, activities.CreateOCIArtifactRequest{
			OwnerID:       run.ID,
			OwnerType:     "install_action_workflow_runs",
			Outputs:       outputs,
			ReuseExisting: job.DemandEnabled(ctx),
		}, &workflow.ActivityOptions{
			RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 1},
		}); err != nil {
			l.Warn("unable to record mirrored action image artifact", zap.Error(err))
		}
	}

	l.Info("bound image-backed action to mirrored manifest",
		zap.String("image_digest_ref", digestRef))

	return digestRef, nil
}

// resolveMirroredDigestRef reads the digest-pinned image ref the oci-sync job
// recorded and verifies it actually carries a digest, so execution can only
// ever run the manifest that was just mirrored.
func resolveMirroredDigestRef(ctx workflow.Context, syncJobID string) (string, state.OCIArtifactOutputs, error) {
	var out state.OCIArtifactOutputs

	syncedJob, err := activities.AwaitGetJobByID(ctx, syncJobID)
	if err != nil {
		return "", out, errors.Wrap(err, "unable to get image sync job to resolve mirrored digest")
	}

	raw, ok := syncedJob.ParsedOutputs["image"]
	if !ok {
		return "", out, errors.New("image sync job recorded no image output, unable to pin action image to a digest")
	}

	if err := mapstructure.Decode(raw, &out); err != nil {
		return "", out, errors.Wrap(err, "unable to decode image sync job output")
	}
	if out.Ref == "" {
		return "", out, errors.New("image sync job recorded an empty image ref, unable to pin action image to a digest")
	}

	parsed, err := reference.Parse(out.Ref)
	if err != nil {
		return "", out, fmt.Errorf("mirrored image ref %q is not a valid reference: %w", out.Ref, err)
	}
	if _, ok := parsed.(reference.Digested); !ok {
		return "", out, fmt.Errorf("mirrored image ref %q is not digest-pinned", out.Ref)
	}

	return out.Ref, out, nil
}

// supportedImageActionPlatform gates image-backed actions to VM runners with a
// host Docker launcher. GCP containers receive the same metadata identity and
// service-account impersonation hints as host-based GCP actions.
func supportedImageActionPlatform(p app.AppRunnerType) bool {
	switch p {
	case app.AppRunnerTypeAWS, app.AppRunnerTypeAzure, app.AppRunnerTypeGCP, app.AppRunnerTypeLocal:
		return true
	default:
		return false
	}
}

// parseActionImageSource splits an app-authored image ref (e.g.
// ghcr.io/acme/tools:v1) into the source registry descriptor and tag the
// oci-sync copier pulls from.
func parseActionImageSource(sourceImage string) (*configs.OCIRegistryRepository, string, error) {
	named, err := reference.ParseDockerRef(sourceImage)
	if err != nil {
		return nil, "", fmt.Errorf("invalid image reference %q: %w", sourceImage, err)
	}

	// Prefer a pinned digest over a tag so a digest-pinned ref actually mirrors
	// and runs the pinned content instead of silently resolving to "latest" (or
	// discarding the digest on a tag+digest ref). oras resolves the source ref
	// by digest or tag, so passing the digest here is valid.
	ref := "latest"
	if digested, ok := named.(reference.Digested); ok {
		ref = digested.Digest().String()
	} else if tagged, ok := named.(reference.Tagged); ok {
		ref = tagged.Tag()
	}

	loginServer := ""
	if reference.Domain(named) == "docker.io" {
		loginServer = "docker.io"
	}

	return &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypePublicOCI,
		Repository:   named.Name(),
		LoginServer:  loginServer,
	}, ref, nil
}

func (s *Signal) updateActionRunStatus(ctx workflow.Context, runID string, status app.InstallActionWorkflowRunStatus, msg string) {
	l := workflow.GetLogger(ctx)

	if err := activities.AwaitUpdateInstallWorkflowRunStatus(ctx, activities.UpdateInstallWorkflowRunStatusRequest{
		RunID:             runID,
		Status:            status,
		StatusDescription: msg,
	}); err != nil {
		l.Error("unable to update run status",
			zap.String("run-id", runID),
			zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateInstallWorkflowRunStatusV2(ctx, statusactivities.UpdateInstallWorkflowRunStatusV2Request{
		RunID:             runID,
		Status:            status,
		StatusDescription: msg,
	}); err != nil {
		l.Error("unable to update run status v2",
			zap.String("run-id", runID),
			zap.Error(err))
	}
}
