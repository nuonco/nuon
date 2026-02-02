// operationroles implements various rules around what role to use for a particular operation
package operationroles

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type EntityOperationRoleMap map[app.OperationType]string

func EntityOperationRoleMapFromHstore(hstore map[string]*string) EntityOperationRoleMap {
	if hstore == nil {
		return nil
	}

	result := make(EntityOperationRoleMap, len(hstore))
	for key, value := range hstore {
		if value != nil {
			result[app.OperationType(key)] = *value
		}
	}
	return result
}

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
	MatrixRules []*app.OperationRoleRule
	// DefaultRole is the role selected if none of the rules assiciate with the pricipal and operation
	DefaultRole string

	StackOutputs *app.InstallStackOutputs

	AppConfig *app.AppConfig
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
func SelectRole(ctx *SelectionContext) (*RoleSelection, error) {
	if ctx == nil {
		return nil, fmt.Errorf("selection context is required")
	}

	// 1. Runtime override (highest precedence)
	if ctx.RuntimeRole != "" {
		roleARN, err := resolveRoleARN(
			ctx.RuntimeRole,
			ctx.AppConfig,
			ctx.StackOutputs,
		)
		if err != nil {
			return nil, fmt.Errorf("runtime role %q: %w", ctx.RuntimeRole, err)
		}
		return &RoleSelection{
			RoleName: ctx.RuntimeRole,
			RoleARN:  roleARN,
			Source:   RoleSelectionSourceRuntime,
		}, nil
	}

	// 2. Entity-level config
	if roleName := findEntityRole(ctx.EntityRoles, ctx.Operation); roleName != "" {
		roleARN, err := resolveRoleARN(roleName, ctx.AppConfig, ctx.StackOutputs)
		if err != nil {
			return nil, fmt.Errorf("entity role %q: %w", roleName, err)
		}
		return &RoleSelection{
			RoleName: roleName,
			RoleARN:  roleARN,
			Source:   RoleSelectionSourceEntity,
		}, nil
	}

	// 3. Matrix rules
	if roleName, found := findMatrixRole(
		ctx.MatrixRules,
		ctx.PrincipalType,
		ctx.PrincipalName,
		ctx.Operation); found {
		roleARN, err := resolveRoleARN(roleName, ctx.AppConfig, ctx.StackOutputs)
		if err != nil {
			return nil, fmt.Errorf("matrix role %q: %w", roleName, err)
		}
		return &RoleSelection{
			RoleName: roleName,
			RoleARN:  roleARN,
			Source:   RoleSelectionSourceMatrix,
		}, nil
	}

	roleARN, err := resolveRoleARN(
		ctx.DefaultRole,
		ctx.AppConfig,
		ctx.StackOutputs,
	)
	if err != nil {
		return nil, fmt.Errorf("default role %q: %w", ctx.DefaultRole, err)
	}

	return &RoleSelection{
		RoleName: ctx.DefaultRole,
		RoleARN:  roleARN,
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
	rules []*app.OperationRoleRule,
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

func resolveRoleARN(roleName string, appCfg *app.AppConfig, stackOutputs *app.InstallStackOutputs) (string, error) {
	// todo(sk): implement this for azure as well
	if stackOutputs == nil || stackOutputs.AWSStackOutputs == nil {
		return "", fmt.Errorf("no AWS stack outputs available")
	}

	availableRoles := make(map[string]string)

	for _, role := range appCfg.PermissionsConfig.CustomRoles {
		if arn, exists := stackOutputs.AWSStackOutputs.CustomRoleARNs[role.Name]; exists {
			availableRoles[role.Name] = arn
		}
	}
	for _, role := range appCfg.BreakGlassConfig.Roles {
		if arn, exists := stackOutputs.AWSStackOutputs.BreakGlassRoleARNs[role.Name]; exists {
			availableRoles[role.Name] = arn
		}
	}

	availableRoles[appCfg.PermissionsConfig.ProvisionRole.Name] = stackOutputs.AWSStackOutputs.ProvisionIAMRoleARN
	availableRoles[appCfg.PermissionsConfig.DeprovisionRole.Name] = stackOutputs.AWSStackOutputs.DeprovisionIAMRoleARN
	availableRoles[appCfg.PermissionsConfig.MaintenanceRole.Name] = stackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN

	roleARN, ok := availableRoles[roleName]
	if !ok {
		return "", fmt.Errorf("role %q not found in install stack outputs", roleName)
	}

	return roleARN, nil
}
