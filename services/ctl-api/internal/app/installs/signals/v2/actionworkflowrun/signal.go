package actionworkflowrun

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
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "install-action-workflow-run"

type Signal struct {
	signal.Hooks
	InstallID               string
	InstallWorkflowID       string
	WorkflowStepID          string
	InstallActionWorkflowID string
	AdhocActionRunID        string
	TriggerType             app.ActionWorkflowTriggerType
	TriggeredByID           string
	TriggeredByType         string
	RunEnvVars              map[string]string
}

var _ signal.Signal = &Signal{}
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithInit = (*Signal)(nil)

func (s *Signal) Init(_ workflow.Context) error {
	s.Hooks.InstallID = &s.InstallID
	s.Hooks.Operation = "action-workflow-run"
	return nil
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
	})
	if err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   actionWorkflowRun.ID,
			StepTargetType: "install_action_workflow_runs", // plugins.TableName would require db instance
		}); err != nil {
			return errors.Wrap(err, "unable to update install action workflow")
		}
	}

	if err := s.executeActionWorkflowRun(ctx, installActionWorkflow.InstallID, actionWorkflowRun.ID); err != nil {
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

	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   run.ID,
			StepTargetType: "install_action_workflow_runs",
		}); err != nil {
			return errors.Wrap(err, "unable to update install action workflow")
		}
	}

	return s.executeActionWorkflowRun(ctx, run.InstallID, run.ID)
}

func (s *Signal) executeActionWorkflowRun(ctx workflow.Context, installID, actionWorkflowRunID string) error {
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

	s.Hooks.LogStreamID = ls.ID

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

	// save runner job plan
	planJSON, err := json.Marshal(planResponse.Plan)
	if err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return errors.Wrap(err, "unable to convert plan to json")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:          runnerJob.ID,
		PlanJSON:       string(planJSON),
		CompositePlan:  plantypes.CompositePlan{ActionWorkflowRunPlan: planResponse.Plan},
		PermissionInfo: operationroles.NewPermissionInfo(planResponse.RoleSelection),
	}); err != nil {
		s.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to save job plan")
		return errors.Wrap(err, "unable to save runner job plan")
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

	if _, err := state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       installID,
		TriggeredByID:   actionWorkflowRunID,
		TriggeredByType: "install_action_workflow_runs", // plugins.TableName would require db instance
	}); err != nil {
		l.Warn("unable to generate state", zap.Error(err))
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
