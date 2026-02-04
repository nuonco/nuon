// operationroles implements various rules around what role to use for a particular operation
package operationroles

import (
	"errors"
	"fmt"

	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"go.uber.org/zap"
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
	MatrixRules []*app.OperationRoleRule
	// DefaultRole is the role selected if none of the rules assiciate with the pricipal and operation
	DefaultRole string
	// Break Glass role
	BreakGlassRole string

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

	// early exit for azure since architecturally azure uses single tenant<> sub id combination
	if ctx.StackOutputs.AzureStackOutputs != nil {
		return &RoleSelection{
			// in case of azure this will be empty, till we figureout azure role based permissions
			RoleName: "azure-placeholder-name",
			RoleARN:  "azure-placeholder-arn",
			Source:   RoleSelectionSourceDefault,
		}, nil
	}

	defaultRoleARN, err := resolveRoleARN(
		ctx.DefaultRole,
		ctx.AppConfig,
		ctx.StackOutputs,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve default role %q: %w", ctx.DefaultRole, err)
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

	// 1.1 for testing purposes we keep sandbox mode at lower priority so we can select roles on runtime
	if ctx.SandboxMode {
		return &RoleSelection{
			RoleName: ctx.DefaultRole,
			// replace this with role arn
			RoleARN: defaultRoleARN,
			Source:  RoleSelectionSourceDefault,
		}, nil
	}

	// 2. Break glass situation, we should respect break glass role definition
	if ctx.BreakGlassRole != "" {
		roleARN, err := resolveRoleARN(ctx.BreakGlassRole, ctx.AppConfig, ctx.StackOutputs)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve break glass role %q: %w", ctx.BreakGlassRole, err)
		}
		return &RoleSelection{
			RoleName: ctx.BreakGlassRole,
			RoleARN:  roleARN,
			Source:   RoleSelectionSourceBreakGlass,
		}, nil
	}

	// 3. Entity-level config
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

	// 4. Matrix rules
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

	return &RoleSelection{
		RoleName: ctx.DefaultRole,
		RoleARN:  defaultRoleARN,
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
	if stackOutputs.AWSStackOutputs != nil {
		availableRoles = getAWSRoleMap(appCfg, stackOutputs.AWSStackOutputs)
	} else if stackOutputs.AzureStackOutputs != nil {
		availableRoles = getAzureRoleMap(appCfg, stackOutputs.AzureStackOutputs)
	} else {
		return "", errors.New("Install stack output nil for both aws and azure")
	}

	roleARN, ok := availableRoles[roleName]
	if !ok {
		return "", fmt.Errorf("role %q not found in install stack outputs", roleName)
	}

	return roleARN, nil
}

func getAWSRoleMap(appCfg *app.AppConfig, stackOutputs *app.AWSStackOutputs) map[string]string {
	availableRoles := make(map[string]string)
	for _, role := range appCfg.PermissionsConfig.CustomRoles {
		if arn, exists := stackOutputs.CustomRoleARNs[role.Name]; exists {
			availableRoles[role.Name] = arn
		}
	}
	for _, role := range appCfg.BreakGlassConfig.Roles {
		if arn, exists := stackOutputs.BreakGlassRoleARNs[role.Name]; exists {
			availableRoles[role.Name] = arn
		}
	}
	availableRoles[appCfg.PermissionsConfig.ProvisionRole.Name] = stackOutputs.ProvisionIAMRoleARN
	availableRoles[appCfg.PermissionsConfig.DeprovisionRole.Name] = stackOutputs.DeprovisionIAMRoleARN
	availableRoles[appCfg.PermissionsConfig.MaintenanceRole.Name] = stackOutputs.MaintenanceIAMRoleARN

	return availableRoles
}

func getAzureRoleMap(appCfg *app.AppConfig, stackOutputs *app.AzureStackOutputs) map[string]string {
	availableRoles := make(map[string]string)
	return availableRoles
}
