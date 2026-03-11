package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func TestGetRoleForAction(t *testing.T) {
	p := &Planner{}
	l := zap.NewNop()

	tests := []struct {
		name               string
		actionName         string
		actionConfigRole   string // Entity-level role from ActionWorkflowConfig
		breakGlassRoleARN  string // Simple string - converted to NullString when building test data
		runtimeRole        string // Runtime role from run.Role
		matrixRules        []*app.AppOperationRoleRule
		useAzure           bool
		expectedOperation  app.OperationType
		expectedRoleSource operationroles.RoleSelectionSource
		expectedRoleName   string
		description        string
	}{
		// No operation rules
		{
			name:               "no_rules_aws",
			actionName:         "deploy-action",
			actionConfigRole:   "",
			breakGlassRoleARN:  "",
			runtimeRole:        "",
			matrixRules:        nil,
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Action with no rules should use default maintenance role on AWS",
		},
		{
			name:               "no_rules_azure",
			actionName:         "deploy-action",
			actionConfigRole:   "",
			breakGlassRoleARN:  "",
			runtimeRole:        "",
			matrixRules:        nil,
			useAzure:           true,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "azure-placeholder-name",
			description:        "Action with no rules on Azure should use Azure placeholder",
		},

		// Entity-level role (ActionWorkflowConfig.Role)
		{
			name:               "entity_role_overrides_default",
			actionName:         "maintenance-action",
			actionConfigRole:   "CustomMaintenanceRole",
			breakGlassRoleARN:  "",
			runtimeRole:        "",
			matrixRules:        nil,
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "CustomMaintenanceRole",
			description:        "Entity role should override default maintenance role",
		},

		// Matrix rules for actions
		{
			name:              "matrix_rule_specific_action",
			actionName:        "deploy-action",
			actionConfigRole:  "",
			breakGlassRoleARN: "",
			runtimeRole:       "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTrigger,
					PrincipalType: "action",
					PrincipalName: "deploy-action",
					Role:          "MatrixDeployActionRole",
				},
			},
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "MatrixDeployActionRole",
			description:        "Action should use matrix rule matching specific action name",
		},
		{
			name:              "matrix_rule_wildcard",
			actionName:        "any-action",
			actionConfigRole:  "",
			breakGlassRoleARN: "",
			runtimeRole:       "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTrigger,
					PrincipalType: "action",
					PrincipalName: "*",
					Role:          "MatrixWildcardRole",
				},
			},
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "MatrixWildcardRole",
			description:        "Action should use matrix wildcard rule",
		},
		{
			name:              "matrix_rule_specific_over_wildcard",
			actionName:        "deploy-action",
			actionConfigRole:  "",
			breakGlassRoleARN: "",
			runtimeRole:       "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTrigger,
					PrincipalType: "action",
					PrincipalName: "deploy-action",
					Role:          "SpecificActionRole",
				},
				{
					Operation:     app.OperationTrigger,
					PrincipalType: "action",
					PrincipalName: "*",
					Role:          "WildcardRole",
				},
			},
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "SpecificActionRole",
			description:        "Specific matrix rule should take precedence over wildcard",
		},
		// Both entity and matrix rules
		{
			name:              "entity_role_overrides_matrix",
			actionName:        "deploy-action",
			actionConfigRole:  "ActionConfigRole",
			breakGlassRoleARN: "",
			runtimeRole:       "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTrigger,
					PrincipalType: "action",
					PrincipalName: "deploy-action",
					Role:          "MatrixRole",
				},
			},
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "ActionConfigRole",
			description:        "Entity role should take precedence over matrix rule (entity > matrix)",
		},

		// Break glass role
		{
			name:               "break_glass_overrides_entity",
			actionName:         "emergency-action",
			actionConfigRole:   "ActionConfigRole",
			breakGlassRoleARN:  "EmergencyRole",
			runtimeRole:        "",
			matrixRules:        nil,
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceBreakGlass,
			expectedRoleName:   "EmergencyRole",
			description:        "Break glass role should override entity role (breakglass > entity)",
		},
		{
			name:              "break_glass_overrides_matrix",
			actionName:        "emergency-action",
			actionConfigRole:  "",
			breakGlassRoleARN: "EmergencyRole",
			runtimeRole:       "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTrigger,
					PrincipalType: "action",
					PrincipalName: "emergency-action",
					Role:          "MatrixRole",
				},
			},
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceBreakGlass,
			expectedRoleName:   "EmergencyRole",
			description:        "Break glass role should override matrix rule (breakglass > matrix)",
		},
		{
			name:              "break_glass_all_rules_present",
			actionName:        "emergency-action",
			actionConfigRole:  "ActionConfigRole",
			breakGlassRoleARN: "EmergencyRole",
			runtimeRole:       "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTrigger,
					PrincipalType: "action",
					PrincipalName: "emergency-action",
					Role:          "MatrixRole",
				},
			},
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceBreakGlass,
			expectedRoleName:   "EmergencyRole",
			description:        "Break glass should override both entity and matrix when all present",
		},

		// Runtime role (highest precedence)
		{
			name:              "runtime_role_overrides_all",
			actionName:        "deploy-action",
			actionConfigRole:  "ActionConfigRole",
			breakGlassRoleARN: "EmergencyRole",
			runtimeRole:       "RuntimeRole",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTrigger,
					PrincipalType: "action",
					PrincipalName: "deploy-action",
					Role:          "MatrixRole",
				},
			},
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceRuntime,
			expectedRoleName:   "RuntimeRole",
			description:        "Runtime role should override all other sources (runtime > breakglass > entity > matrix > default)",
		},
		{
			name:               "runtime_role_only",
			actionName:         "deploy-action",
			actionConfigRole:   "",
			breakGlassRoleARN:  "",
			runtimeRole:        "RuntimeRole",
			matrixRules:        nil,
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceRuntime,
			expectedRoleName:   "RuntimeRole",
			description:        "Runtime role should work even when no other rules exist",
		},
		{
			name:               "runtime_role_overrides_break_glass",
			actionName:         "emergency-action",
			actionConfigRole:   "",
			breakGlassRoleARN:  "EmergencyRole",
			runtimeRole:        "RuntimeRole",
			matrixRules:        nil,
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceRuntime,
			expectedRoleName:   "RuntimeRole",
			description:        "Runtime role should override break glass role",
		},
		{
			name:               "runtime_role_on_azure",
			actionName:         "deploy-action",
			actionConfigRole:   "",
			breakGlassRoleARN:  "",
			runtimeRole:        "RuntimeRole",
			matrixRules:        nil,
			useAzure:           true,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "azure-placeholder-name",
			description:        "Azure should return placeholder regardless of runtime role",
		},
		{
			name:               "entity_role_missing_fallback_to_default",
			actionName:         "deploy-action",
			actionConfigRole:   "MissingEntityRole", // Not in stack outputs
			breakGlassRoleARN:  "",
			runtimeRole:        "",
			matrixRules:        nil,
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Should fallback to default when entity role is missing from stack outputs",
		},
		{
			name:              "matrix_role_missing_fallback_to_default",
			actionName:        "deploy-action",
			actionConfigRole:  "",
			breakGlassRoleARN: "",
			runtimeRole:       "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTrigger,
					PrincipalType: "action",
					PrincipalName: "deploy-action",
					Role:          "MissingMatrixRole", // Not in stack outputs
				},
			},
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Should fallback to default when matrix role is missing from stack outputs",
		},
		{
			name:               "runtime_role_missing_fallback_to_default",
			actionName:         "deploy-action",
			actionConfigRole:   "",
			breakGlassRoleARN:  "",
			runtimeRole:        "MissingRuntimeRole", // Not in stack outputs
			matrixRules:        nil,
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Should fallback to default when runtime role is missing from stack outputs",
		},
		{
			name:               "break_glass_role_missing_fallback_to_default",
			actionName:         "emergency-action",
			actionConfigRole:   "",
			breakGlassRoleARN:  "MissingBreakGlassRole", // Not in stack outputs
			runtimeRole:        "",
			matrixRules:        nil,
			useAzure:           false,
			expectedOperation:  app.OperationTrigger,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Should fallback to default when break glass role is missing from stack outputs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appConfig := &app.AppConfig{
				PermissionsConfig: app.AppPermissionsConfig{
					ProvisionRole: app.AppAWSIAMRoleConfig{
						Name: "ProvisionRole",
					},
					MaintenanceRole: app.AppAWSIAMRoleConfig{
						Name: "MaintenanceRole",
					},
					DeprovisionRole: app.AppAWSIAMRoleConfig{
						Name: "DeprovisionRole",
					},
					CustomRoles: []app.AppAWSIAMRoleConfig{
						{Name: "ActionConfigRole"},
						{Name: "CustomMaintenanceRole"},
						{Name: "MatrixDeployActionRole"},
						{Name: "MatrixWildcardRole"},
						{Name: "SpecificActionRole"},
						{Name: "WildcardRole"},
						{Name: "ComponentRole"},
						{Name: "DeployRole"},
						{Name: "DeployActionRole"},
						{Name: "MatrixRole"},
						{Name: "RuntimeRole"},
					},
				},
				BreakGlassConfig: app.AppBreakGlassConfig{
					Roles: []app.AppAWSIAMRoleConfig{
						{Name: "EmergencyRole"},
					},
				},
				OperationRoleConfig: app.AppOperationRoleConfig{
					Rules: tt.matrixRules,
				},
			}

			run := &app.InstallActionWorkflowRun{
				Role: tt.runtimeRole,
				ActionWorkflowConfig: app.ActionWorkflowConfig{
					Role:              tt.actionConfigRole,
					BreakGlassRoleARN: generics.NewNullString(tt.breakGlassRoleARN),
					ActionWorkflow: app.ActionWorkflow{
						Name: tt.actionName,
					},
				},
			}

			stack := &app.InstallStack{
				InstallStackOutputs: app.InstallStackOutputs{},
			}

			if tt.useAzure {
				stack.InstallStackOutputs.AzureStackOutputs = &app.AzureStackOutputs{
					SubscriptionID: "test-subscription-id",
				}
			} else {
				stack.InstallStackOutputs.AWSStackOutputs = &app.AWSStackOutputs{
					Region:                "us-west-2",
					ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/ProvisionRole",
					MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/MaintenanceRole",
					DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/DeprovisionRole",
					CustomRoleARNs: map[string]string{
						"ActionConfigRole":       "arn:aws:iam::123456789012:role/ActionConfigRole",
						"CustomMaintenanceRole":  "arn:aws:iam::123456789012:role/CustomMaintenanceRole",
						"MatrixDeployActionRole": "arn:aws:iam::123456789012:role/MatrixDeployActionRole",
						"MatrixWildcardRole":     "arn:aws:iam::123456789012:role/MatrixWildcardRole",
						"SpecificActionRole":     "arn:aws:iam::123456789012:role/SpecificActionRole",
						"WildcardRole":           "arn:aws:iam::123456789012:role/WildcardRole",
						"ComponentRole":          "arn:aws:iam::123456789012:role/ComponentRole",
						"DeployRole":             "arn:aws:iam::123456789012:role/DeployRole",
						"DeployActionRole":       "arn:aws:iam::123456789012:role/DeployActionRole",
						"MatrixRole":             "arn:aws:iam::123456789012:role/MatrixRole",
						"RuntimeRole":            "arn:aws:iam::123456789012:role/RuntimeRole",
					},
					BreakGlassRoleARNs: map[string]string{
						"EmergencyRole": "arn:aws:iam::123456789012:role/EmergencyRole",
					},
				}
			}

			installState := &state.State{
				ID:   "test-install",
				Name: "test-install",
			}

			roleSelection, operation, err := p.getRoleForAction(
				l,
				appConfig,
				run,
				stack,
				installState,
			)

			require.NoError(t, err, "getRoleForAction should not return error for test: %s", tt.description)
			assert.Equal(t, tt.expectedOperation, operation, "Operation type mismatch: %s", tt.description)
			assert.Equal(t, tt.expectedRoleSource, roleSelection.Source, "Role source mismatch: %s", tt.description)
			assert.Equal(t, tt.expectedRoleName, roleSelection.RoleName, "Role name mismatch: %s", tt.description)
			assert.NotEmpty(t, roleSelection.RoleARN, "Role ARN should be populated: %s", tt.description)

			// For Azure, we expect placeholder ARN
			if tt.useAzure {
				assert.Equal(t, "azure-placeholder-arn", roleSelection.RoleARN, "Azure should use placeholder ARN: %s", tt.description)
			} else {
				assert.Contains(t, roleSelection.RoleARN, tt.expectedRoleName, "Role ARN should contain role name: %s", tt.description)
			}

			t.Logf("%s: Got role=%s (source=%s, operation=%s)",
				tt.name,
				roleSelection.RoleName,
				roleSelection.Source,
				operation,
			)
		})
	}
}
