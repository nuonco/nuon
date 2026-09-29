package operationroles

import (
	"fmt"
	"maps"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SelectionContext struct {
	Operation app.OperationType

	PrincipalType principal.Type
	PrincipalName string

	RuntimeRole    string
	EntityRoles    EntityOperationRoleMap
	MatrixRules    []*app.AppOperationRoleRule
	DefaultRole    string
	BreakGlassRole string

	StackOutputs *app.InstallStackOutputs

	AppConfig *app.AppConfig

	InstallState *state.State
}

type RoleSelectionSource string

const (
	RoleSelectionSourceRuntime    RoleSelectionSource = "runtime"
	RoleSelectionSourceEntity     RoleSelectionSource = "entity"
	RoleSelectionSourceMatrix     RoleSelectionSource = "matrix"
	RoleSelectionSourceDefault    RoleSelectionSource = "default"
	RoleSelectionSourceBreakGlass RoleSelectionSource = "breakglass"
)

type RoleSelection struct {
	RoleName           string                           `temporaljson:"role_name"`
	UnrenderedRoleName string                           `temporaljson:"unrendered_role_name"`
	RoleARN            string                           `temporaljson:"role_arn"`
	Source             RoleSelectionSource              `temporaljson:"source"`
	Trace              []app.InstallRoleSelectionRecord `temporaljson:"trace"`
}

type SelectionError struct {
	Err   error
	Trace []app.InstallRoleSelectionRecord
}

func (e *SelectionError) Error() string { return e.Err.Error() }
func (e *SelectionError) Unwrap() error { return e.Err }

func SelectRole(ctx *SelectionContext, l *zap.Logger) (*RoleSelection, error) {
	if ctx == nil {
		return nil, fmt.Errorf("selection context is required")
	}

	selection, err := selectRole(ctx)
	if err != nil {
		return nil, err
	}

	renderedRoleName, err := renderRoleName(selection.RoleName, ctx.InstallState)
	if err != nil {
		return nil, &SelectionError{
			Err:   errors.Wrap(err, "unable to render default role name"),
			Trace: selection.Trace,
		}
	}

	roleARN, err := resolveRoleARN(
		renderedRoleName,
		ctx.AppConfig,
		ctx.StackOutputs,
		ctx.InstallState,
	)
	if err != nil {
		return nil, &SelectionError{
			Err:   fmt.Errorf("unable to resolve role ARN for %q: %w", renderedRoleName, err),
			Trace: selection.Trace,
		}
	}

	selection.RoleARN = roleARN

	return selection, nil
}

func SelectDefaultRole(ctx *SelectionContext) (*RoleSelection, error) {
	if ctx.DefaultRole == "" {
		return nil, fmt.Errorf("no default role configured for %s", ctx.Operation)
	}

	renderedDefaultRole, err := renderRoleName(ctx.DefaultRole, ctx.InstallState)
	if err != nil {
		return nil, fmt.Errorf("unable to render default role name: %w", err)
	}
	roleARN, err := resolveRoleARN(
		renderedDefaultRole,
		ctx.AppConfig,
		ctx.StackOutputs,
		ctx.InstallState,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve role ARN for %q: %w", renderedDefaultRole, err)
	}

	return &RoleSelection{
		RoleName:           renderedDefaultRole,
		UnrenderedRoleName: ctx.DefaultRole,
		Source:             RoleSelectionSourceDefault,
		RoleARN:            roleARN,
		Trace: []app.InstallRoleSelectionRecord{
			{
				RoleName:   renderedDefaultRole,
				RoleSource: string(RoleSelectionSourceDefault),
				Available:  true,
				Selected:   true,
			},
		},
	}, nil
}

func selectRole(ctx *SelectionContext) (*RoleSelection, error) {
	if ctx.StackOutputs == nil {
		return nil, fmt.Errorf("stack outputs are required")
	}

	var trace []app.InstallRoleSelectionRecord

	if ctx.DefaultRole == "" {
		return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("no default role configured for %s", ctx.Operation)}
	}

	renderedDefaultRole, err := renderRoleName(ctx.DefaultRole, ctx.InstallState)
	if err != nil {
		return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("unable to render default role name: %w", err)}
	}

	runtimeAvailable := ctx.RuntimeRole != ""
	if runtimeAvailable {
		renderedRuntimeRole, err := renderRoleName(ctx.RuntimeRole, ctx.InstallState)
		if err != nil {
			return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("unable to render runtime role name: %w", err)}
		}
		trace = append(trace, app.InstallRoleSelectionRecord{
			RoleName:   renderedRuntimeRole,
			RoleSource: string(RoleSelectionSourceRuntime),
			Available:  true,
			Selected:   true,
		})
		return &RoleSelection{
			RoleName:           renderedRuntimeRole,
			UnrenderedRoleName: ctx.RuntimeRole,
			Source:             RoleSelectionSourceRuntime,
			Trace:              trace,
		}, nil
	}
	trace = append(trace, app.InstallRoleSelectionRecord{
		RoleName:   ctx.RuntimeRole,
		RoleSource: string(RoleSelectionSourceRuntime),
		Available:  false,
	})

	breakGlassAvailable := ctx.BreakGlassRole != ""
	if breakGlassAvailable {
		renderedBreakGlassRole, err := renderRoleName(ctx.BreakGlassRole, ctx.InstallState)
		if err != nil {
			return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("unable to render break glass role name: %w", err)}
		}
		trace = append(trace, app.InstallRoleSelectionRecord{
			RoleName:   renderedBreakGlassRole,
			RoleSource: string(RoleSelectionSourceBreakGlass),
			Available:  true,
			Selected:   true,
		})
		return &RoleSelection{
			RoleName:           renderedBreakGlassRole,
			UnrenderedRoleName: ctx.BreakGlassRole,
			Source:             RoleSelectionSourceBreakGlass,
			Trace:              trace,
		}, nil
	}
	trace = append(trace, app.InstallRoleSelectionRecord{
		RoleName:   ctx.BreakGlassRole,
		RoleSource: string(RoleSelectionSourceBreakGlass),
		Available:  false,
	})

	entityRoleName := findEntityRole(ctx.EntityRoles, ctx.Operation)
	entityAvailable := entityRoleName != ""
	if entityAvailable {
		renderedEntityRole, err := renderRoleName(entityRoleName, ctx.InstallState)
		if err != nil {
			return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("unable to render entity role name: %w", err)}
		}
		trace = append(trace, app.InstallRoleSelectionRecord{
			RoleName:   renderedEntityRole,
			RoleSource: string(RoleSelectionSourceEntity),
			Available:  true,
			Selected:   true,
		})
		return &RoleSelection{
			RoleName:           renderedEntityRole,
			UnrenderedRoleName: entityRoleName,
			Source:             RoleSelectionSourceEntity,
			Trace:              trace,
		}, nil
	}
	trace = append(trace, app.InstallRoleSelectionRecord{
		RoleName:   entityRoleName,
		RoleSource: string(RoleSelectionSourceEntity),
		Available:  false,
	})

	matrixRoleName, matrixFound, err := findMatrixRole(
		ctx.MatrixRules,
		ctx.PrincipalType,
		ctx.PrincipalName,
		ctx.Operation,
		ctx.InstallState)
	if err != nil {
		return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("unable to evaluate matrix rules: %w", err)}
	}
	if matrixFound {
		renderedMatrixRole, err := renderRoleName(matrixRoleName, ctx.InstallState)
		if err != nil {
			return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("unable to render matrix role name: %w", err)}
		}
		trace = append(trace, app.InstallRoleSelectionRecord{
			RoleName:   renderedMatrixRole,
			RoleSource: string(RoleSelectionSourceMatrix),
			Available:  true,
			Selected:   true,
		})
		return &RoleSelection{
			RoleName:           renderedMatrixRole,
			UnrenderedRoleName: matrixRoleName,
			Source:             RoleSelectionSourceMatrix,
			Trace:              trace,
		}, nil
	}
	trace = append(trace, app.InstallRoleSelectionRecord{
		RoleName:   matrixRoleName,
		RoleSource: string(RoleSelectionSourceMatrix),
		Available:  false,
	})

	trace = append(trace, app.InstallRoleSelectionRecord{
		RoleName:   renderedDefaultRole,
		RoleSource: string(RoleSelectionSourceDefault),
		Available:  true,
		Selected:   true,
	})

	return &RoleSelection{
		RoleName:           renderedDefaultRole,
		UnrenderedRoleName: ctx.DefaultRole,
		Source:             RoleSelectionSourceDefault,
		Trace:              trace,
	}, nil
}

func findEntityRole(roles EntityOperationRoleMap, operation app.OperationType) string {
	if roles == nil {
		return ""
	}
	roleName, ok := roles[operation]
	if !ok {
		return ""
	}
	return roleName
}

func findMatrixRole(
	rules []*app.AppOperationRoleRule,
	principalType principal.Type,
	principalName string,
	operation app.OperationType,
	installState *state.State,
) (string, bool, error) {
	renderedPrincipalName, err := renderRoleName(principalName, installState)
	if err != nil {
		return "", false, fmt.Errorf("unable to render principal name %q: %w", principalName, err)
	}

	for _, rule := range rules {
		if rule.Operation != operation {
			continue
		}

		if rule.PrincipalType != principalType {
			continue
		}

		switch principalType {
		case principal.TypeComponent, principal.TypeAction:
			if rule.PrincipalName == "*" {
				return rule.Role, true, nil
			}
			renderedRulePrincipalName, err := renderRoleName(rule.PrincipalName, installState)
			if err != nil {
				return "", false, fmt.Errorf("unable to render rule principal name %q: %w", rule.PrincipalName, err)
			}
			if renderedRulePrincipalName == renderedPrincipalName {
				return rule.Role, true, nil
			}
		case principal.TypeSandbox:
			if rule.PrincipalName == "" {
				return rule.Role, true, nil
			}
		}
	}

	return "", false, nil
}

func renderRoleName(roleName string, installState *state.State) (string, error) {
	if installState == nil || roleName == "" {
		return roleName, nil
	}

	stateMap, err := installState.AsMap()
	if err != nil {
		return roleName, fmt.Errorf("unable to convert install state to map: %w", err)
	}

	rendered, err := render.RenderV2(roleName, stateMap)
	if err != nil {
		return roleName, fmt.Errorf("unable to render role name template %q: %w", roleName, err)
	}

	return rendered, nil
}

func resolveRoleARN(
	renderedRoleName string,
	appCfg *app.AppConfig,
	installStackOutputs *app.InstallStackOutputs,
	installState *state.State,
) (string, error) {
	if installStackOutputs == nil {
		return "", fmt.Errorf("stack outputs are required")
	}

	var stackOutput app.StackOutput
	if installStackOutputs.AzureStackOutputs != nil {
		stackOutput = installStackOutputs.AzureStackOutputs
	} else if installStackOutputs.AWSStackOutputs != nil {
		stackOutput = installStackOutputs.AWSStackOutputs
	} else if installStackOutputs.GCPStackOutputs != nil {
		stackOutput = installStackOutputs.GCPStackOutputs
	} else {
		return "", errors.New("stack outputs must have either AWS, Azure, or GCP outputs")
	}

	availableRoles, err := getRoleMap(appCfg, stackOutput, installState)
	if err != nil {
		return "", fmt.Errorf("unable to get AWS role map: %w", err)
	}

	roleARN, ok := availableRoles[renderedRoleName]
	if !ok {
		return "", fmt.Errorf("role %s not found in install stack outputs, please enable it in install stack", renderedRoleName)
	}

	return roleARN, nil
}

func getRoleMap(appCfg *app.AppConfig, stackOutputs app.StackOutput, installState *state.State) (map[string]string, error) {
	if appCfg == nil {
		return nil, fmt.Errorf("app config is required")
	}
	if installState == nil {
		return nil, fmt.Errorf("install state is required for role rendering")
	}

	availableRoles := make(map[string]string)

	customRoles, err := stackOutputs.CustomRoles()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch customRoles role %w", err)
	}
	maps.Copy(availableRoles, customRoles)

	breakGlassRoles, err := stackOutputs.BreakGlassRoles()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch break glass role %w", err)
	}
	maps.Copy(availableRoles, breakGlassRoles)

	stateMap, err := installState.AsMap()
	if err != nil {
		return nil, fmt.Errorf("unable to convert install state to map: %w", err)
	}

	standardRoles := []struct {
		name   string
		idFunc func() (string, error)
	}{
		{appCfg.PermissionsConfig.ProvisionRole.Name, stackOutputs.ProvisionRoleID},
		{appCfg.PermissionsConfig.DeprovisionRole.Name, stackOutputs.DeprovisionRoleID},
		{appCfg.PermissionsConfig.MaintenanceRole.Name, stackOutputs.MaintenanceRoleID},
	}

	for _, r := range standardRoles {
		rendered, err := render.RenderV2(r.name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render role name template %q: %w", r.name, err)
		}
		roleID, err := r.idFunc()
		if err != nil {
			return nil, fmt.Errorf("unable to fetch role ID for %q: %w", r.name, err)
		}
		availableRoles[rendered] = roleID
	}

	return availableRoles, nil
}
