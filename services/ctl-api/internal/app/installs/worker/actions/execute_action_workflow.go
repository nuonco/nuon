package actions

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	installstate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

// @temporal-gen workflow
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

// @temporal-gen workflow
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
	runPlan, err := plan.AwaitCreateActionWorkflowRunPlan(ctx, &plan.CreateActionRunPlanRequest{
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
	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return errors.Wrap(err, "unable to convert plan to json")
	}

	appConfig, err := activities.AwaitGetAppConfigByID(ctx, run.Install.AppConfigID)
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to get app config")
		return fmt.Errorf("unable to get app config: %w", err)
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, run.Install.ID)
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to get install stack")
		return errors.Wrap(err, "unable to get install stack")
	}

	// Get install state for role name rendering
	installState, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: run.Install.ID,
	})
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to get install state")
		return fmt.Errorf("unable to get install state: %w", err)
	}

	compositePlan := plantypes.CompositePlan{
		ActionWorkflowRunPlan: runPlan,
	}

	roleSelection, operation, err := w.getRoleForAction(l, appConfig, run, stack, installState)
	if err != nil {
		l.Error("unable to evaluate role for action operation", zap.Error(err))
		w.updateActionRunStatus(
			ctx,
			run.ID,
			app.InstallActionRunStatusError,
			"unable to select role",
		)
		return errors.Wrap(err, "unable to evaluate role for action")
	}

	l.Info("selected role for action workflow",
		zap.String("role_name", roleSelection.RoleName),
		zap.String("role_arn", roleSelection.RoleARN),
		zap.String("source", string(roleSelection.Source)),
		zap.String("action_workflow", run.ActionWorkflowConfig.ActionWorkflow.Name),
		zap.String("operation", string(operation)),
	)

	planAuth, err := plan.CreatePlanAuth(
		stack.InstallStackOutputs,
		roleSelection.RoleARN,
		fmt.Sprintf("install-action-workflow-%s", run.ID),
	)
	if err != nil {
		l.Error("unable to build plan auth for action operation", zap.Error(err))
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create auth config")
		return errors.Wrap(err, "unable to create plan auth")
	}
	compositePlan.Auth = planAuth

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: compositePlan,
	}); err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to save job plan")
		return errors.Wrap(err, "unable to save runner job plan")
	}

	planJSON = nil
	runPlan = nil

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

func (w *Workflows) getRoleForAction(
	l *zap.Logger,
	appConfig *app.AppConfig,
	run *app.InstallActionWorkflowRun,
	stack *app.InstallStack,
	installState *state.State,
) (*operationroles.RoleSelection, app.OperationType, error) {
	operation := app.OperationTrigger

	var entityRoles map[app.OperationType]string
	if run.ActionWorkflowConfig.Role != "" {
		entityRoles = map[app.OperationType]string{
			operation: run.ActionWorkflowConfig.Role,
		}
	}

	var defaultRole string
	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		defaultRole = appConfig.PermissionsConfig.MaintenanceRole.Name
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		defaultRole = "azure-maintainence-mock-role-name"
	default:
	}

	var breakGlassRole string
	if run.ActionWorkflowConfig.BreakGlassRoleARN.Valid {
		breakGlassRole = run.ActionWorkflowConfig.BreakGlassRoleARN.String
	}

	selectionCtx := &operationroles.SelectionContext{
		Operation:      operation,
		PrincipalType:  principal.TypeAction,
		PrincipalName:  run.ActionWorkflowConfig.ActionWorkflow.Name,
		RuntimeRole:    run.Role,
		EntityRoles:    entityRoles,
		MatrixRules:    appConfig.OperationRoleConfig.Rules,
		DefaultRole:    defaultRole,
		AppConfig:      appConfig,
		StackOutputs:   &stack.InstallStackOutputs,
		BreakGlassRole: breakGlassRole,
		InstallState:   installState,
	}

	roleSelection, err := operationroles.SelectRole(selectionCtx, l)
	if err != nil {
		l.Warn("dynamic role selection failed, falling back to default role",
			zap.Error(err),
			zap.String("default_role", selectionCtx.DefaultRole),
		)

		var fallbackErr error
		roleSelection, fallbackErr = operationroles.GetDefaultRoleSelection(selectionCtx)
		if fallbackErr != nil {
			return nil, "", fmt.Errorf("unable to get default role: %w", fallbackErr)
		}

		l.Warn("using default role for action",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
		)
	}

	return roleSelection, operation, nil
}
