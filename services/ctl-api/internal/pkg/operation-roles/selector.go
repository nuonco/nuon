// operationroles implements various rules around what role to use for a particular operation
package operationroles

import (
	"fmt"
	"maps"

	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// SelectionContext contains all information needed for role selection
type SelectionContext struct {
	Operation app.OperationType

	// "component", "sandbox", "action"
	PrincipalType principal.Type
	// Component/action name (empty for sandbox)
	PrincipalName string

	// Configuration sources (in precedence order)
	// --role flag from CLI/UI (highest precedence)
	RuntimeRole string
	// Component/sandbox/action config
	EntityRoles EntityOperationRoleMap
	// App-level rules from DB
	MatrixRules []*app.AppOperationRoleRule
	// DefaultRole is the role selected if none of the rules associate with the principal and operation
	DefaultRole string
	// Break Glass role
	BreakGlassRole string

	StackOutputs *app.InstallStackOutputs

	AppConfig *app.AppConfig

	// Install state for rendering role names with templating
	InstallState *state.State
}

// RoleSelectionSource represents where a role selection came from
type RoleSelectionSource string

const (
	// selected at runtime
	RoleSelectionSourceRuntime RoleSelectionSource = "runtime"
	// defined in entity definition, in component, action or sandbox
	RoleSelectionSourceEntity RoleSelectionSource = "entity"
	// defined in app config rules
	RoleSelectionSourceMatrix RoleSelectionSource = "matrix"
	// existing behavior
	RoleSelectionSourceDefault RoleSelectionSource = "default"
	// break glass
	RoleSelectionSourceBreakGlass RoleSelectionSource = "breakglass"
)

type RoleSelection struct {
	RoleName           string `temporaljson:"role_name"`
	UnrenderedRoleName string `temporaljson:"unrendered_role_name"`
	// RoleArn is arn/id/unique identifier for the role depending on cloud provider
	RoleARN string                               `temporaljson:"role_arn"`
	Source  RoleSelectionSource                  `temporaljson:"source"`
	Trace   []app.RunnerJobPermissionTraceRecord `temporaljson:"trace"`
}

type SelectionError struct {
	Err   error
	Trace []app.RunnerJobPermissionTraceRecord
}

func (e *SelectionError) Error() string { return e.Err.Error() }
func (e *SelectionError) Unwrap() error { return e.Err }

// SelectRole determines which role to use based on precedence rules
// Precedence (highest to lowest):
// 1. Runtime override (CLI --role flag or UI selection)
// 2. Entity-level config (component/sandbox/action specific)
// 3. Matrix rules (app-level operation_roles config)
// 4. Default roles (provision/maintenance/deprovision)
func SelectRole(ctx *SelectionContext, l *zap.Logger) (*RoleSelection, error) {
	if ctx == nil {
		return nil, fmt.Errorf("selection context is required")
	}

	selection, err := selectRole(ctx)
	if err != nil {
		return nil, err
	}

	// If RoleARN is already set (e.g., Azure placeholder), return as-is
	if selection.RoleARN != "" {
		return selection, nil
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
		Trace: []app.RunnerJobPermissionTraceRecord{
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
	// Add nil check for StackOutputs before dereferencing
	if ctx.StackOutputs == nil {
		return nil, fmt.Errorf("stack outputs are required")
	}

	var trace []app.RunnerJobPermissionTraceRecord

	// early exit for azure since architecturally azure uses single tenant<> sub id combination
	if ctx.StackOutputs.AzureStackOutputs != nil {
		trace = append(trace, app.RunnerJobPermissionTraceRecord{
			RoleName:   "azure-placeholder-name",
			RoleSource: string(RoleSelectionSourceDefault),
			RoleID:     "azure-placeholder-arn",
			Available:  true,
			Selected:   true,
		})
		return &RoleSelection{
			// in case of azure this will be empty, till we figureout azure role based permissions
			RoleName: "azure-placeholder-name",
			RoleARN:  "azure-placeholder-arn",
			Source:   RoleSelectionSourceDefault,
			Trace:    trace,
		}, nil
	}

	if ctx.DefaultRole == "" {
		return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("no default role configured for %s", ctx.Operation)}
	}

	// Render default role name with install state (fixes template rendering in sandbox and default paths)
	renderedDefaultRole, err := renderRoleName(ctx.DefaultRole, ctx.InstallState)
	if err != nil {
		return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("unable to render default role name: %w", err)}
	}

	// 1. Runtime override (highest precedence)
	runtimeAvailable := ctx.RuntimeRole != ""
	if runtimeAvailable {
		renderedRuntimeRole, err := renderRoleName(ctx.RuntimeRole, ctx.InstallState)
		if err != nil {
			return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("unable to render runtime role name: %w", err)}
		}
		trace = append(trace, app.RunnerJobPermissionTraceRecord{
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
	trace = append(trace, app.RunnerJobPermissionTraceRecord{
		RoleName:   ctx.RuntimeRole,
		RoleSource: string(RoleSelectionSourceRuntime),
		Available:  false,
	})

	// 2. Break glass situation, we should respect break glass role definition
	breakGlassAvailable := ctx.BreakGlassRole != ""
	if breakGlassAvailable {
		renderedBreakGlassRole, err := renderRoleName(ctx.BreakGlassRole, ctx.InstallState)
		if err != nil {
			return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("unable to render break glass role name: %w", err)}
		}
		trace = append(trace, app.RunnerJobPermissionTraceRecord{
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
	trace = append(trace, app.RunnerJobPermissionTraceRecord{
		RoleName:   ctx.BreakGlassRole,
		RoleSource: string(RoleSelectionSourceBreakGlass),
		Available:  false,
	})

	// 3. Entity-level config
	entityRoleName := findEntityRole(ctx.EntityRoles, ctx.Operation)
	entityAvailable := entityRoleName != ""
	if entityAvailable {
		renderedEntityRole, err := renderRoleName(entityRoleName, ctx.InstallState)
		if err != nil {
			return nil, &SelectionError{Trace: trace, Err: fmt.Errorf("unable to render entity role name: %w", err)}
		}
		trace = append(trace, app.RunnerJobPermissionTraceRecord{
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
	trace = append(trace, app.RunnerJobPermissionTraceRecord{
		RoleName:   entityRoleName,
		RoleSource: string(RoleSelectionSourceEntity),
		Available:  false,
	})

	// 4. Matrix rules
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
		trace = append(trace, app.RunnerJobPermissionTraceRecord{
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
	trace = append(trace, app.RunnerJobPermissionTraceRecord{
		RoleName:   matrixRoleName,
		RoleSource: string(RoleSelectionSourceMatrix),
		Available:  false,
	})

	trace = append(trace, app.RunnerJobPermissionTraceRecord{
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

	// Find matching rule
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

// renderRoleName renders a role name template using install state
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

// ResolveRoleARN looks up the ARN for a given role name from stack outputs.
// Currently mostly does heavy lifting for AWS since Azure is not yet supported.
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
		return "", nil
	} else if installStackOutputs.AWSStackOutputs != nil {
		// Try AWS second
		stackOutput = installStackOutputs.AWSStackOutputs
	} else if installStackOutputs.GCPStackOutputs != nil {
		// Try GCP third
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

// getRoleMap returns a map of renderRoleName role name to role arn
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
