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
	// under sandbox mode make sure to choose either provision deprovision or maintenance
	SandboxMode bool

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
	// DefaultRole is the role selected if none of the rules assiciate with the pricipal and operation
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
	RoleName string
	RoleARN  string
	Source   RoleSelectionSource
}

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
		return nil, errors.WithMessagef(err, "unable to select role for input context")
	}

	// If RoleARN is already set (e.g., Azure placeholder), return as-is
	if selection.RoleARN != "" {
		return selection, nil
	}

	renderedRoleName, err := renderRoleName(selection.RoleName, ctx.InstallState)
	if err != nil {
		return nil, errors.Wrap(err, "unable to render default role name")
	}

	roleARN, err := resolveRoleARN(renderedRoleName, ctx.AppConfig, ctx.StackOutputs, ctx.InstallState)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve role ARN for %q: %w", renderedRoleName, err)
	}

	selection.RoleARN = roleARN

	return selection, nil
}

func GetDefaultRoleSelection(ctx *SelectionContext) (*RoleSelection, error) {
	if ctx.DefaultRole == "" {
		return nil, fmt.Errorf("no default role configured for %s", ctx.Operation)
	}

	renderedDefaultRole, err := renderRoleName(ctx.DefaultRole, ctx.InstallState)
	if err != nil {
		return nil, fmt.Errorf("unable to render default role name: %w", err)
	}
	roleARN, err := resolveRoleARN(renderedDefaultRole, ctx.AppConfig, ctx.StackOutputs, ctx.InstallState)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve role ARN for %q: %w", renderedDefaultRole, err)
	}

	return &RoleSelection{
		RoleName: renderedDefaultRole,
		Source:   RoleSelectionSourceDefault,
		RoleARN:  roleARN,
	}, nil
}

func selectRole(ctx *SelectionContext) (*RoleSelection, error) {
	// Add nil check for StackOutputs before dereferencing
	if ctx.StackOutputs == nil {
		return nil, fmt.Errorf("stack outputs are required")
	}

	// early exit for azure since architecturally azure uses single tenant<> sub id combination
	if ctx.StackOutputs.AzureStackOutputs != nil {
		return &RoleSelection{
			// in case of azure this will be empty, till we figureout azure role based permissions
			RoleName: "azure-placeholder-name",
			RoleARN:  "azure-placeholder-arn",
			Source:   RoleSelectionSourceDefault,
		}, nil
	}

	if ctx.DefaultRole == "" {
		return nil, fmt.Errorf("no default role configured for %s", ctx.Operation)
	}

	// Render default role name with install state (fixes template rendering in sandbox and default paths)
	renderedDefaultRole, err := renderRoleName(ctx.DefaultRole, ctx.InstallState)
	if err != nil {
		return nil, fmt.Errorf("unable to render default role name: %w", err)
	}

	// 1. Runtime override (highest precedence)
	if ctx.RuntimeRole != "" {
		renderedRuntimeRole, err := renderRoleName(ctx.RuntimeRole, ctx.InstallState)
		if err != nil {
			return nil, fmt.Errorf("unable to render runtime role name: %w", err)
		}
		return &RoleSelection{
			RoleName: renderedRuntimeRole,
			Source:   RoleSelectionSourceRuntime,
		}, nil
	}

	// 1.1 for testing purposes we keep sandbox mode at lower priority so we can select roles on runtime
	if ctx.SandboxMode {
		return &RoleSelection{
			RoleName: renderedDefaultRole,
			Source:   RoleSelectionSourceDefault,
		}, nil
	}

	// 2. Break glass situation, we should respect break glass role definition
	if ctx.BreakGlassRole != "" {
		renderedBreakGlassRole, err := renderRoleName(ctx.BreakGlassRole, ctx.InstallState)
		if err != nil {
			return nil, fmt.Errorf("unable to render break glass role name: %w", err)
		}
		return &RoleSelection{
			RoleName: renderedBreakGlassRole,
			Source:   RoleSelectionSourceBreakGlass,
		}, nil
	}

	// 3. Entity-level config
	if roleName := findEntityRole(ctx.EntityRoles, ctx.Operation); roleName != "" {
		renderedEntityRole, err := renderRoleName(roleName, ctx.InstallState)
		if err != nil {
			return nil, fmt.Errorf("unable to render entity role name: %w", err)
		}
		return &RoleSelection{
			RoleName: renderedEntityRole,
			Source:   RoleSelectionSourceEntity,
		}, nil
	}

	// 4. Matrix rules
	if roleName, found := findMatrixRole(
		ctx.MatrixRules,
		ctx.PrincipalType,
		ctx.PrincipalName,
		ctx.Operation); found {
		renderedMatrixRole, err := renderRoleName(roleName, ctx.InstallState)
		if err != nil {
			return nil, fmt.Errorf("unable to render matrix role name: %w", err)
		}
		return &RoleSelection{
			RoleName: renderedMatrixRole,
			Source:   RoleSelectionSourceMatrix,
		}, nil
	}

	return &RoleSelection{
		RoleName: renderedDefaultRole,
		Source:   RoleSelectionSourceDefault,
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
) (string, bool) {
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
			if rule.PrincipalName == principalName || rule.PrincipalName == "*" {
				return rule.Role, true
			}
		case principal.TypeSandbox:
			if rule.PrincipalName == "" {
				return rule.Role, true
			}
		}
	}

	return "", false
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
func resolveRoleARN(renderedRoleName string, appCfg *app.AppConfig, stackOutputs *app.InstallStackOutputs, installState *state.State) (string, error) {
	if stackOutputs == nil {
		return "", fmt.Errorf("stack outputs are required")
	}

	var availableRoles map[string]string
	var err error

	// Try Azure first
	if stackOutputs.AzureStackOutputs != nil {
		availableRoles = getAzureRoleMap(appCfg, stackOutputs.AzureStackOutputs)
	} else if stackOutputs.AWSStackOutputs != nil {
		// Try AWS second
		availableRoles, err = getAWSRoleMap(appCfg, stackOutputs.AWSStackOutputs, installState)
		if err != nil {
			return "", fmt.Errorf("unable to get AWS role map: %w", err)
		}
	} else if stackOutputs.GCPStackOutputs != nil {
		// Try GCP third
		availableRoles, err = getGCPSAMap(appCfg, stackOutputs.GCPStackOutputs, installState)
		if err != nil {
			return "", fmt.Errorf("unable to get GCP SA map: %w", err)
		}
	} else {
		return "", errors.New("stack outputs must have either AWS, Azure, or GCP outputs")
	}

	roleARN, ok := availableRoles[renderedRoleName]
	if !ok {
		return "", fmt.Errorf("role %s not found in install stack outputs, please enable it in install stack", renderedRoleName)
	}

	return roleARN, nil
}

// getAWSRoleMAP returns a map of renderRoleName role name to role arn
func getAWSRoleMap(appCfg *app.AppConfig, stackOutputs *app.AWSStackOutputs, installState *state.State) (map[string]string, error) {
	if appCfg == nil {
		return nil, fmt.Errorf("app config is required")
	}
	if installState == nil {
		return nil, fmt.Errorf("install state is required for role rendering")
	}

	availableRoles := make(map[string]string)

	maps.Copy(availableRoles, stackOutputs.CustomRoleARNs)
	maps.Copy(availableRoles, stackOutputs.BreakGlassRoleARNs)
	stateMap, err := installState.AsMap()
	if err != nil {
		return nil, fmt.Errorf("unable to convert install state to map: %w", err)
	}

	renderedProvisionRoleName, err := render.RenderV2(appCfg.PermissionsConfig.ProvisionRole.Name, stateMap)
	if err != nil {
		return nil, fmt.Errorf("unable to render role name template %q: %w", appCfg.PermissionsConfig.ProvisionRole.Name, err)
	}
	availableRoles[renderedProvisionRoleName] = stackOutputs.ProvisionIAMRoleARN

	renderedDeprovisionRoleName, err := render.RenderV2(appCfg.PermissionsConfig.DeprovisionRole.Name, stateMap)
	if err != nil {
		return nil, fmt.Errorf("unable to render role name template %q: %w", appCfg.PermissionsConfig.DeprovisionRole.Name, err)
	}
	availableRoles[renderedDeprovisionRoleName] = stackOutputs.DeprovisionIAMRoleARN

	renderedMaintenanceRoleName, err := render.RenderV2(appCfg.PermissionsConfig.MaintenanceRole.Name, stateMap)
	if err != nil {
		return nil, fmt.Errorf("unable to render role name template %q: %w", appCfg.PermissionsConfig.MaintenanceRole.Name, err)
	}
	availableRoles[renderedMaintenanceRoleName] = stackOutputs.MaintenanceIAMRoleARN

	return availableRoles, nil
}

func getAzureRoleMap(appCfg *app.AppConfig, stackOutputs *app.AzureStackOutputs) map[string]string {
	availableRoles := make(map[string]string)
	return availableRoles
}

// getGCPSAMap returns a map of rendered role name to GCP service account email.
func getGCPSAMap(appCfg *app.AppConfig, stackOutputs *app.GCPStackOutputs, installState *state.State) (map[string]string, error) {
	if appCfg == nil {
		return nil, fmt.Errorf("app config is required")
	}
	if installState == nil {
		return nil, fmt.Errorf("install state is required for role rendering")
	}

	stateMap, err := installState.AsMap()
	if err != nil {
		return nil, fmt.Errorf("unable to convert install state to map: %w", err)
	}

	availableRoles := make(map[string]string)

	renderedProvisionRoleName, err := render.RenderV2(appCfg.PermissionsConfig.ProvisionRole.Name, stateMap)
	if err != nil {
		return nil, fmt.Errorf("unable to render provision role name: %w", err)
	}
	availableRoles[renderedProvisionRoleName] = stackOutputs.ProvisionSAEmail

	renderedMaintenanceRoleName, err := render.RenderV2(appCfg.PermissionsConfig.MaintenanceRole.Name, stateMap)
	if err != nil {
		return nil, fmt.Errorf("unable to render maintenance role name: %w", err)
	}
	availableRoles[renderedMaintenanceRoleName] = stackOutputs.MaintenanceSAEmail

	renderedDeprovisionRoleName, err := render.RenderV2(appCfg.PermissionsConfig.DeprovisionRole.Name, stateMap)
	if err != nil {
		return nil, fmt.Errorf("unable to render deprovision role name: %w", err)
	}
	availableRoles[renderedDeprovisionRoleName] = stackOutputs.DeprovisionSAEmail

	if appCfg.PermissionsConfig.BreakGlassRole.Name != "" {
		renderedBreakGlassRoleName, err := render.RenderV2(appCfg.PermissionsConfig.BreakGlassRole.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render break glass role name: %w", err)
		}
		availableRoles[renderedBreakGlassRoleName] = stackOutputs.BreakGlassSAEmail
	}

	return availableRoles, nil
}
