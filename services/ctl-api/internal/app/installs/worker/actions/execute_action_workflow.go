package actions

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	installstate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

// @temporal-gen-v2 workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ExecuteActionWorkflowRun(ctx workflow.Context, req signals.RequestSignal) error {
	if req.ActionWorkflowRunID == "" {
		return errors.New("action workflow run ID is required")
	}

	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, req.ActionWorkflowRunID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	if req.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         req.WorkflowStepID,
			StepTargetID:   run.ID,
			StepTargetType: plugins.TableName(w.db, run),
		}); err != nil {
			return errors.Wrap(err, "unable to update workflow step target")
		}
	}

	if err := w.executeActionWorkflowRun(ctx, run.InstallID, run.ID); err != nil {
		return errors.Wrap(err, "unable to execute action workflow run")
	}

	return nil
}

// @temporal-gen-v2 workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ExecuteActionWorkflow(ctx workflow.Context, req signals.RequestSignal) error {
	installActionWorkflow, err := activities.AwaitGetInstallActionWorkflowByID(ctx, req.InstallActionWorkflowTrigger.InstallActionWorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get install action workflow")
	}

	install, err := activities.AwaitGetByInstallID(ctx, installActionWorkflow.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	// Skip cron-triggered actions if sandbox is not active to prevent workflow history size from growing.
	if req.InstallActionWorkflowTrigger.TriggerType == app.ActionWorkflowTriggerTypeCron {
		switch install.SandboxStatus {
		// We may want to add more cases here in the future.
		case app.InstallSandboxStatusProvisioning,
			app.InstallSandboxStatusDeprovisioning,
			app.InstallSandboxStatusDeprovisioned:
			l, _ := log.WorkflowLogger(ctx)
			if l != nil {
				l.Info("skipping cron action execution - sandbox not active",
					zap.String("install_id", installActionWorkflow.InstallID),
					zap.String("action_workflow_name", installActionWorkflow.ActionWorkflow.Name),
					zap.String("sandbox_status", string(install.SandboxStatus)),
				)
			}
			return nil // Skip this cron execution
		}
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
		InstallWorkflowID:       req.InstallWorkflowID,
		TriggerType:             req.InstallActionWorkflowTrigger.TriggerType,
		TriggeredByID:           req.InstallActionWorkflowTrigger.TriggeredByID,
		TriggeredByType:         req.InstallActionWorkflowTrigger.TriggeredByType,
		RunEnvVars:              generics.ToPtrStringMap(req.InstallActionWorkflowTrigger.RunEnvVars),
		Role:                    req.InstallActionWorkflowTrigger.Role,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	if req.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         req.WorkflowStepID,
			StepTargetID:   actionWorkflowRun.ID,
			StepTargetType: plugins.TableName(w.db, actionWorkflowRun),
		}); err != nil {
			return errors.Wrap(err, "unable to update install action workflow")
		}
	}

	if err := w.executeActionWorkflowRun(ctx, installActionWorkflow.InstallID, actionWorkflowRun.ID); err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	return nil
}

func (w *Workflows) executeActionWorkflowRun(ctx workflow.Context, installID, actionWorkflowRunID string) error {
	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, actionWorkflowRunID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	l, err := log.WorkflowLogger(ctx)
	if err == nil {
		l.Warn("creating a new logger for executing action")
	}
	parentLS, _ := cctx.GetLogStreamWorkflow(ctx)

	lsReq := activities.CreateLogStreamRequest{
		ActionWorkflowRunID: actionWorkflowRunID,
	}
	if parentLS != nil {
		lsReq.ParentLogStreamID = parentLS.ID
	}
	w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusInProgress, "in-progress")
	ls, err := activities.AwaitCreateLogStream(ctx, lsReq)
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}

	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, ls.ID)
	}()
	ctx = cctx.SetLogStreamWorkflowContext(ctx, ls)

	l, err = log.WorkflowLogger(ctx)
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create log stream")
		return errors.Wrap(err, "unable to set log stream on context")
	}

	ls, err = cctx.GetLogStreamWorkflow(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get log stream")
	}

	l.Info("creating plan for executing action run")
	planResponse, err := plan.AwaitCreateActionWorkflowRunPlan(ctx, &plan.CreateActionRunPlanRequest{
		ActionWorkflowRunID: actionWorkflowRunID,
		WorkflowID:          fmt.Sprintf("%s-create-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create plan")
		return errors.Wrap(err, "unable to create plan")
	}

	// execute job
	l.Info("creating runner job to execute action")
	metadata := map[string]string{
		"install_id":             installID,
		"action_workflow_run_id": run.ID,
	}

	if run.ActionWorkflowConfigID.Valid {
		metadata["action_workflow_name"] = run.ActionWorkflowConfig.ActionWorkflow.Name
		metadata["action_workflow_id"] = run.ActionWorkflowConfig.ActionWorkflowID
	} else {
		actionName := "Adhoc action"
		if len(run.Steps) > 0 && run.Steps[0].AdHocConfig != nil && run.Steps[0].AdHocConfig.Name != "" {
			actionName = run.Steps[0].AdHocConfig.Name
		}
		metadata["action_workflow_name"] = actionName
		metadata["action_workflow_id"] = ""
	}

	runnerJob, err := activities.AwaitCreateActionWorkflowRunRunnerJob(ctx, &activities.CreateActionWorkflowRunRunnerJob{
		ActionWorkflowRunID: actionWorkflowRunID,
		RunnerID:            run.Install.RunnerID,
		LogStreamID:         ls.ID,
		Metadata:            metadata,
	})
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return errors.Wrap(err, "unable to create runner job")
	}

	// save runner job plan
	planJSON, err := json.Marshal(planResponse.Plan)
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return errors.Wrap(err, "unable to convert plan to json")
	}

	compositePlan := plantypes.CompositePlan{
		ActionWorkflowRunPlan: planResponse.Plan,
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: compositePlan,
	}); err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to save job plan")
		return errors.Wrap(err, "unable to save runner job plan")
	}

	if err := activities.AwaitRecordInstallRoleUsage(ctx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     installID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: planResponse.RoleSelection,
	}); err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to record install role usage")
		return errors.Wrap(err, "unable to record install role usage")
	}

	planJSON = nil

	// now queue and execute the job
	l.Info("executing runner job")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   run.Install.RunnerID,
		JobID:      runnerJob.ID,
		WorkflowID: "actions-install-run-exec-job" + run.ID,
	})
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "job failed")
		return errors.Wrap(err, "runner job failed")
	}

	w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusFinished, "finished")

	_, err = installstate.AwaitGenerateState(ctx, &installstate.GenerateStateRequest{
		InstallID:       installID,
		TriggeredByID:   actionWorkflowRunID,
		TriggeredByType: plugins.TableName(w.db, run),
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return nil
}
