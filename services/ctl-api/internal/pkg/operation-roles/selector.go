// operationroles implements various rules around what role to use for a particular operation
package operationroles

import (
	"fmt"
	"slices"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// SelectionContext contains all information needed for role selection
type SelectionContext struct {
	Operation config.OperationType

	// "component", "sandbox", "action"
	PrincipalType config.PrincipalType
	// Component/action name (empty for sandbox)
	PrincipalName string

	// Configuration sources (in precedence order)
	// --role flag from CLI/UI (highest precedence)
	RuntimeRole string
	// Component/sandbox/action config
	EntityRoles []*config.EntityOperationRole
	// App-level rules from DB
	MatrixRules []*app.OperationRoleRule
	// For role ARN resolution
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

var DefaultRoles = map[config.OperationType]string{
	config.OperationProvision:   "provision",
	config.OperationDeprovision: "deprovision",
	config.OperationUpdate:      "maintenance",
	config.OperationReprovision: "provision",
	config.OperationTrigger:     "maintenance",
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

	// 4. Default role
	defaultRole, ok := DefaultRoles[ctx.Operation]
	if !ok {
		return nil, fmt.Errorf("no default role found for operation %s", ctx.Operation)
	}

	roleARN, err := resolveRoleARN(
		defaultRole,
		ctx.AppConfig,
		ctx.StackOutputs,
	)
	if err != nil {
		return nil, fmt.Errorf("default role %q: %w", defaultRole, err)
	}

	return &RoleSelection{
		RoleName: defaultRole,
		RoleARN:  roleARN,
		Source:   RoleSelectionSourceDefault,
	}, nil
}

func findEntityRole(roles []*config.EntityOperationRole, operation config.OperationType) string {
	for _, r := range roles {
		if r.Operation == operation {
			return r.RoleName
		}
	}
	return ""
}

func findMatrixRole(
	rules []*app.OperationRoleRule,
	principalType config.PrincipalType,
	principalName string,
	operation config.OperationType,
) (string, bool) {
	var principals []string

	switch principalType {
	case config.PrincipalTypeComponent:
		principals = []string{
			fmt.Sprintf("nuon::component:%s", principalName),
			"nuon::component:*",
		}
	case config.PrincipalTypeSandbox:
		principals = []string{"nuon::sandbox"}
	case config.PrincipalTypeAction:
		principals = []string{
			fmt.Sprintf("nuon::action:%s", principalName),
			"nuon::action:*",
		}
	}

	// Find matching rule
	for _, rule := range rules {
		if rule.Operation == app.OperationType(operation) || slices.Contains(principals, rule.Principal) {
			return rule.Role, true
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
