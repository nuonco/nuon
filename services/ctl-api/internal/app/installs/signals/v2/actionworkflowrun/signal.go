package actionworkflowrun

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
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

var _ signal.Signal = &Signal{}
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)
var _ signal.SignalWithAutoRetry = (*Signal)(nil)
var _ signal.SignalWithMaxRetries = (*Signal)(nil)
var _ signal.SignalWithMaxAutoRetries = (*Signal)(nil)
var _ signal.SignalWithOnRetry = (*Signal)(nil)

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	if s.runnerJobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, s.runnerJobID)
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

	install, err := activities.AwaitGetByInstallID(ctx, installActionWorkflow.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}

	found := false
	for _, workflowCfg := range appCfg.ActionWorkflowConfigs {
		if workflowCfg.ActionWorkflowID == installActionWorkflow.ActionWorkflowID {
			found = true
			break
		}
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

	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			s.updateActionRunStatus(updateCtx, actionWorkflowRun.ID, app.InstallActionRunStatusCancelled, "action workflow run cancelled")
		}
	}()

	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   actionWorkflowRun.ID,
			StepTargetType: "install_action_workflow_runs", // plugins.TableName would require db instance
		}); err != nil {
			return errors.Wrap(err, "unable to update install action workflow")
		}
	}

	if err := s.executeActionWorkflowRun(ctx, install, actionWorkflowRun.ID); err != nil {
		return errors.Wrap(err, "unable to execute action workflow run")
	}

	return nil
}

func (s *Signal) executeAdhocRun(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)
	l.Info("executing adhoc action workflow run signal",
		zap.String("adhoc_action_run_id", s.AdhocActionRunID))

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

	return s.executeActionWorkflowRun(ctx, install, run.ID)
}

func (s *Signal) executeActionWorkflowRun(ctx workflow.Context, install *app.Install, actionWorkflowRunID string) error {
	installID := install.ID
	l := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, actionWorkflowRunID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	parentLS, _ := cctx.GetLogStreamWorkflow(ctx)

	lsReq := activities.CreateLogStreamRequest{
		ActionWorkflowRunID: actionWorkflowRunID,
	}
	if parentLS != nil {
		lsReq.ParentLogStreamID = parentLS.ID
	}

	s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusInProgress, "in-progress")

	ls, err := activities.AwaitCreateLogStream(ctx, lsReq)
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}

	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, ls.ID)
	}()

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

	l.Info("creating plan for executing action run")
	planResponse, err := plan.AwaitCreateActionWorkflowRunPlan(ctx, &plan.CreateActionRunPlanRequest{
		ActionWorkflowRunID: actionWorkflowRunID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-create-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create plan")
		return errors.Wrap(err, "unable to create plan")
	}

	// execute job
	l.Info("creating runner job to execute action")
	runnerJob, err := activities.AwaitCreateActionWorkflowRunRunnerJob(ctx, &activities.CreateActionWorkflowRunRunnerJob{
		ActionWorkflowRunID: actionWorkflowRunID,
		RunnerID:            run.Install.RunnerID,
		LogStreamID:         ls.ID,
		Metadata: map[string]string{
			"install_id":             installID,
			"action_workflow_name":   run.ActionWorkflowConfig.ActionWorkflow.Name,
			"action_workflow_run_id": run.ID,
			"action_workflow_id":     run.ActionWorkflowConfig.ActionWorkflowID,
		},
	})
	if err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return errors.Wrap(err, "unable to create runner job")
	}
	s.runnerJobID = runnerJob.ID

	// save runner job plan
	planJSON, err := json.Marshal(planResponse.Plan)
	if err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return errors.Wrap(err, "unable to convert plan to json")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: plantypes.CompositePlan{ActionWorkflowRunPlan: planResponse.Plan},
	}); err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to save job plan")
		return errors.Wrap(err, "unable to save runner job plan")
	}

	if err := activities.AwaitRecordInstallRoleUsage(ctx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     installID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: planResponse.RoleSelection,
	}); err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to record install role usage")
		return errors.Wrap(err, "unable to record install role usage")
	}

	planJSON = nil

	// now queue and execute the job
	l.Info("executing runner job")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID: run.Install.RunnerID,
		JobID:    runnerJob.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: "actions-install-run-exec-job" + run.ID,
	})
	if err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "job failed")
		return errors.Wrap(err, "runner job failed")
	}

	s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusFinished, "finished")

	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	if err := stategen.HintOrGenerate(ctx, stategen.Request{
		StateGenV2:      statemanager.UseStateGenV2(orgEnabled, install.Metadata),
		InstallID:       installID,
		Targets:         statemanager.TargetsForHint(statemanager.HintActionRan, s.InstallActionWorkflowID),
		ForceAll:        true,
		TriggeredByID:   actionWorkflowRunID,
		TriggeredByType: "install_action_workflow_runs",
	}); err != nil {
		return err
	}

	return nil
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
