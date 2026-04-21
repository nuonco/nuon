package operationroles

import (
	"errors"
	"fmt"

	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"go.uber.org/zap"
)

// GetRoleForDeploy selects the role for a component deploy or teardown operation.
func GetRoleForDeploy(
	l *zap.Logger,
	appCfg *app.AppConfig,
	installDeploy *app.InstallDeploy,
	compCfgConn *app.ComponentConfigConnection,
	stack *app.InstallStack,
	installState *state.State,
) (*RoleSelection, app.OperationType, error) {
	operation := app.OperationDeploy
	if installDeploy.Type == app.InstallDeployTypeTeardown {
		operation = app.OperationTeardown
	}

	selectionCtx := &SelectionContext{
		Operation:     operation,
		PrincipalType: principal.TypeComponent,
		PrincipalName: compCfgConn.Component.Name,
		RuntimeRole:   installDeploy.Role,
		EntityRoles: EntityOperationRoleMapFromHstore(
			compCfgConn.OperationRoles,
		),
		MatrixRules:  appCfg.OperationRoleConfig.Rules,
		DefaultRole:  appCfg.PermissionsConfig.MaintenanceRole.Name,
		AppConfig:    appCfg,
		StackOutputs: &stack.InstallStackOutputs,
		InstallState: installState,
	}

	roleSelection, err := SelectRole(selectionCtx, l)
	if err != nil {
		var selErr *SelectionError
		var failedTrace []app.InstallRoleSelectionRecord
		if errors.As(err, &selErr) {
			failedTrace = selErr.Trace
		}

		l.Warn("dynamic role selection failed, falling back to default role",
			zap.Error(err),
			zap.String("default_role", selectionCtx.DefaultRole),
		)

		roleSelection, err = SelectDefaultRole(selectionCtx)
		if err != nil {
			return nil, "", fmt.Errorf("unable to get default role: %w", err)
		}
		roleSelection.Trace = append(failedTrace, roleSelection.Trace...)

		l.Warn("using default role for component deploy",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
		)
	}

	return roleSelection, operation, nil
}

// GetRoleForSandbox selects the role for a sandbox provision/reprovision/deprovision operation.
func GetRoleForSandbox(
	l *zap.Logger,
	appCfg *app.AppConfig,
	run *app.InstallSandboxRun,
	stack *app.InstallStack,
	installState *state.State,
) (*RoleSelection, app.OperationType, error) {
	var operation app.OperationType
	switch run.RunType {
	case app.SandboxRunTypeProvision:
		operation = app.OperationProvision
	case app.SandboxRunTypeReprovision:
		operation = app.OperationReprovision
	case app.SandboxRunTypeDeprovision:
		operation = app.OperationDeprovision
	default:
		operation = app.OperationProvision
	}

	defaultRole := appCfg.PermissionsConfig.ProvisionRole.Name
	if operation == app.OperationDeprovision {
		defaultRole = appCfg.PermissionsConfig.DeprovisionRole.Name
	}

	selectionCtx := &SelectionContext{
		Operation:     operation,
		PrincipalType: principal.TypeSandbox,
		PrincipalName: "",
		RuntimeRole:   run.Role,
		EntityRoles: EntityOperationRoleMapFromHstore(
			appCfg.SandboxConfig.OperationRoles,
		),
		MatrixRules:  appCfg.OperationRoleConfig.Rules,
		DefaultRole:  defaultRole,
		AppConfig:    appCfg,
		StackOutputs: &stack.InstallStackOutputs,
		InstallState: installState,
	}

	roleSelection, err := SelectRole(selectionCtx, l)
	if err != nil {
		var selErr *SelectionError
		var failedTrace []app.InstallRoleSelectionRecord
		if errors.As(err, &selErr) {
			failedTrace = selErr.Trace
		}

		l.Warn("dynamic role selection failed, falling back to default role",
			zap.Error(err),
			zap.String("default_role", selectionCtx.DefaultRole),
		)

		roleSelection, err = SelectDefaultRole(selectionCtx)
		if err != nil {
			return nil, "", fmt.Errorf("unable to get default role: %w", err)
		}
		roleSelection.Trace = append(failedTrace, roleSelection.Trace...)

		l.Warn("using default role for sandbox",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
		)
	}

	return roleSelection, operation, nil
}

// GetRoleForAction selects the role for an action workflow trigger operation.
func GetRoleForAction(
	l *zap.Logger,
	appCfg *app.AppConfig,
	run *app.InstallActionWorkflowRun,
	stack *app.InstallStack,
	installState *state.State,
) (*RoleSelection, app.OperationType, error) {
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

	selectionCtx := &SelectionContext{
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

	roleSelection, err := SelectRole(selectionCtx, l)
	if err != nil {
		var selErr *SelectionError
		var failedTrace []app.InstallRoleSelectionRecord
		if errors.As(err, &selErr) {
			failedTrace = selErr.Trace
		}

		l.Warn("dynamic role selection failed, falling back to default role",
			zap.Error(err),
			zap.String("default_role", selectionCtx.DefaultRole),
		)

		roleSelection, err = SelectDefaultRole(selectionCtx)
		if err != nil {
			l.Error("unable to get default role", zap.Error(err))
			return nil, "", fmt.Errorf("unable to get default role: %w", err)
		}
		roleSelection.Trace = append(failedTrace, roleSelection.Trace...)

		l.Warn("using default role for action",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
		)
	}

	return roleSelection, operation, nil
}
