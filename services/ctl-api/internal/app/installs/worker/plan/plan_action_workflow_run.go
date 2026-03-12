package plan

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func (p *Planner) createActionWorkflowRunPlan(ctx workflow.Context, runID string) (*plantypes.ActionWorkflowRunPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	l.Info("creating plan for executing action workflow")
	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, runID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get run")
	}

	org, err := activities.AwaitGetOrgByInstallID(ctx, run.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install id")
	}

	install, err := activities.AwaitGetByInstallID(ctx, run.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	// step 2 - interpolate all variables in the set
	l.Debug("fetching install state")
	state, err := activities.AwaitGetInstallStateByInstallID(ctx, run.InstallID)
	if err != nil {
		l.Error("unable to get install state", zap.Error(err))
		return nil, errors.Wrap(err, "unable to get install state")
	}

	stateMap, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert state to map")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, run.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	builtInEnvVars, err := p.getBuiltinEnvVars(ctx, run)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get env vars")
	}

	overrideEnvVars, err := p.getOverrideEnvVars(ctx, run)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get override env vars")
	}

	attrs := make(map[string]string, 0)
	if !run.ActionWorkflowConfigID.Empty() {
		attrs["action.name"] = run.ActionWorkflowConfig.ActionWorkflow.Name
		attrs["action.id"] = run.ActionWorkflowConfig.ActionWorkflow.ID
	} else {
		name := generics.FirstNonEmptyString(run.Steps[0].AdHocConfig.Name, "Adhoc Action")
		attrs["action.name"] = name
		attrs["action.id"] = run.ID
	}

	cloudAuth, err := p.getAuthForActionWorkflowRun(ctx, stack.InstallStackOutputs, run, appCfg, stack, state)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get auth for action workflow run")
	}
	var clusterInfo *kube.ClusterInfo
	if run.EnableKubeConfig.Valid && run.EnableKubeConfig.Bool {
		clusterInfo, err = p.getKubeClusterInfo(ctx, stack, state, cloudAuth)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get cluster info")
		}
	}

	plan := &plantypes.ActionWorkflowRunPlan{
		InstallID:       run.InstallID,
		ID:              runID,
		Steps:           make([]*plantypes.ActionWorkflowRunStepPlan, 0),
		BuiltinEnvVars:  builtInEnvVars,
		OverrideEnvVars: overrideEnvVars,
		Attrs:           attrs,
		ClusterInfo:     clusterInfo,
		AzureAuth:       cloudAuth.Azure,
		AWSAuth:         cloudAuth.AWS,
		GCPAuth:         cloudAuth.GCP,
	}

	if !run.ActionWorkflowConfigID.Empty() {
		for idx, stepCfg := range run.Steps {
			l.Debug(fmt.Sprintf("creating plan for step %d", idx))
			stepPlan, err := p.createStepPlan(ctx, &stepCfg, stateMap, run.InstallID)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("unable to create plan for step %d", idx))
			}

			plan.Steps = append(plan.Steps, stepPlan)
		}
	} else {
		stepPlan, err := p.createAdhocStepPlan(ctx, &run.Steps[0], stateMap, run.InstallID)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("unable to create adhoc step plan"))
		}
		plan.Steps = append(plan.Steps, stepPlan)
	}

	if org.SandboxMode {
		targetRefs := helpers.GetActionReferences(appCfg, run.ActionWorkflowConfig.ActionWorkflow.Name)

		plan.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: refs.GetFakeRefs(targetRefs),
		}
	}

	l.Info("successfully created plan")
	return plan, nil
}

// TODO(ja): make this a method on the run struct?
func hstoreToMap(hstore pgtype.Hstore) map[string]string {
	result := make(map[string]string)
	for key, value := range hstore {
		result[key] = *value
	}
	return result
}

func (p *Planner) getRoleForAction(
	l *zap.Logger,
	appCfg *app.AppConfig,
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
		MatrixRules:    appCfg.OperationRoleConfig.Rules,
		DefaultRole:    appCfg.PermissionsConfig.MaintenanceRole.Name,
		AppConfig:      appCfg,
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
			l.Error("unable to get default role", zap.Error(fallbackErr))
			return nil, "", fmt.Errorf("unable to get default role: %w", fallbackErr)
		}

		l.Warn("using default role for action",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
		)
	}

	return roleSelection, operation, nil
}

func (p *Planner) getAuthForActionWorkflowRun(
	ctx workflow.Context,
	outputs app.InstallStackOutputs,
	run *app.InstallActionWorkflowRun,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	installState *state.State,
) (*CloudAuth, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	roleSelection, operation, err := p.getRoleForAction(l, appCfg, run, stack, installState)
	if err != nil {
		return nil, err
	}

	l.Info("selected role for action workflow run plan",
		zap.String("role_name", roleSelection.RoleName),
		zap.String("role_arn", roleSelection.RoleARN),
		zap.String("source", string(roleSelection.Source)),
		zap.String("operation", string(operation)),
		zap.String("run_id", run.ID),
	)

	return getCloudAuth(roleSelection, &outputs, fmt.Sprintf("action-workflow-%s", run.ID))
}
