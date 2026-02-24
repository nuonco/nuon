package components

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetRoleForDeploy(t *testing.T) {
	// Create a workflows instance for testing
	w := &Workflows{}

	tests := []struct {
		name               string
		installDeployType  app.InstallDeployType
		installDeployRole  string // runtime role
		componentRoles     map[app.OperationType]string
		matrixRules        []*app.AppOperationRoleRule
		expectedOperation  app.OperationType
		expectedRoleSource operationroles.RoleSelectionSource
		expectedRoleName   string
		description        string
	}{
		// No operation rules anywhere, runs as always, no change in operation
		{
			name:               "no_rules_deploy_apply",
			installDeployType:  app.InstallDeployTypeApply,
			installDeployRole:  "",
			matrixRules:        nil,
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Deploy with no rules should use default maintenance role",
		},
		{
			name:               "no_rules_teardown",
			installDeployType:  app.InstallDeployTypeTeardown,
			installDeployRole:  "",
			matrixRules:        nil,
			expectedOperation:  app.OperationTeardown,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Deploy with no rules should use default maintenance role",
		},

		// Component has operation rules
		{
			name:              "component_role_deploy",
			installDeployType: app.InstallDeployTypeApply,
			installDeployRole: "",
			componentRoles: map[app.OperationType]string{
				app.OperationDeploy: "ComponentDeployRole",
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "ComponentDeployRole",
			description:        "Deploy should use component-specific deploy role",
		},
		{
			name:              "component_role_teardown",
			installDeployType: app.InstallDeployTypeTeardown,
			installDeployRole: "",
			componentRoles: map[app.OperationType]string{
				app.OperationTeardown: "ComponentTeardownRole",
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationTeardown,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "ComponentTeardownRole",
			description:        "Teardown should use component-specific teardown role",
		},
		{
			name:              "component_has_different_operation_role",
			installDeployType: app.InstallDeployTypeApply,
			installDeployRole: "",
			componentRoles: map[app.OperationType]string{
				app.OperationProvision: "ComponentProvisionRole",
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Deploy should fall back to default when component only has provision role",
		},

		// app config has matrix rules, no component rules
		{
			name:              "matrix_rule_matches_component_deploy",
			installDeployType: app.InstallDeployTypeApply,
			installDeployRole: "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationDeploy,
					PrincipalType: "component",
					PrincipalName: "test-component",
					Role:          "MatrixDeployRole",
				},
			},
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "MatrixDeployRole",
			description:        "Deploy should use matrix rule matching component name",
		},
		{
			name:              "matrix_rule_matches_wildcard",
			installDeployType: app.InstallDeployTypeApply,
			installDeployRole: "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationDeploy,
					PrincipalType: "component",
					PrincipalName: "*",
					Role:          "WildcardDeployRole",
				},
			},
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "WildcardDeployRole",
			description:        "Deploy should use matrix wildcard rule",
		},
		{
			name:              "matrix_rule_teardown",
			installDeployType: app.InstallDeployTypeTeardown,
			installDeployRole: "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTeardown,
					PrincipalType: "component",
					PrincipalName: "*",
					Role:          "MatrixTeardownRole",
				},
			},
			expectedOperation:  app.OperationTeardown,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "MatrixTeardownRole",
			description:        "Teardown should use matrix rule for teardown operation",
		},
		{
			name:              "matrix_rule_no_match",
			installDeployType: app.InstallDeployTypeApply,
			installDeployRole: "",
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTeardown, // Different operation
					PrincipalType: "component",
					PrincipalName: "*",
					Role:          "MatrixTeardownRole",
				},
			},
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Deploy should use default when matrix rule is for different operation",
		},

		// appconfig matrix rules AND component roles

		{
			name:              "component_role_overrides_matrix",
			installDeployType: app.InstallDeployTypeApply,
			installDeployRole: "",
			componentRoles: map[app.OperationType]string{
				app.OperationDeploy: "ComponentDeployRole",
			},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationDeploy,
					PrincipalType: "component",
					PrincipalName: "*",
					Role:          "MatrixDeployRole",
				},
			},
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "ComponentDeployRole",
			description:        "Component role should take precedence over matrix rule (entity > matrix)",
		},
		{
			name:              "both_present_teardown",
			installDeployType: app.InstallDeployTypeTeardown,
			installDeployRole: "",
			componentRoles: map[app.OperationType]string{
				app.OperationTeardown: "ComponentTeardownRole",
			},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTeardown,
					PrincipalType: "component",
					PrincipalName: "test-component",
					Role:          "MatrixTeardownRole",
				},
			},
			expectedOperation:  app.OperationTeardown,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "ComponentTeardownRole",
			description:        "Component teardown role should override matrix rule",
		},

		// Case 5: Runtime role (highest precedence)

		{
			name:              "runtime_role_overrides_all_deploy",
			installDeployType: app.InstallDeployTypeApply,
			installDeployRole: "RuntimeDeployRole",
			componentRoles: map[app.OperationType]string{
				app.OperationDeploy: "ComponentDeployRole",
			},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationDeploy,
					PrincipalType: "component",
					PrincipalName: "*",
					Role:          "MatrixDeployRole",
				},
			},
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceRuntime,
			expectedRoleName:   "RuntimeDeployRole",
			description:        "Runtime role should override all other sources (runtime > entity > matrix > default)",
		},
		{
			name:              "runtime_role_overrides_all_teardown",
			installDeployType: app.InstallDeployTypeTeardown,
			installDeployRole: "RuntimeTeardownRole",
			componentRoles: map[app.OperationType]string{
				app.OperationTeardown: "ComponentTeardownRole",
			},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTeardown,
					PrincipalType: "component",
					PrincipalName: "*",
					Role:          "MatrixTeardownRole",
				},
			},
			expectedOperation:  app.OperationTeardown,
			expectedRoleSource: operationroles.RoleSelectionSourceRuntime,
			expectedRoleName:   "RuntimeTeardownRole",
			description:        "Runtime role should override all sources for teardown",
		},
		{
			name:               "runtime_role_only",
			installDeployType:  app.InstallDeployTypeApply,
			installDeployRole:  "RuntimeOnlyRole",
			matrixRules:        nil,
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceRuntime,
			expectedRoleName:   "RuntimeOnlyRole",
			description:        "Runtime role should work even when no other rules exist",
		},

		// ============================================
		// Case 6: Multiple matrix rules (priority order)
		// ============================================
		{
			name:              "matrix_specific_over_wildcard",
			installDeployType: app.InstallDeployTypeApply,
			installDeployRole: "",
			componentRoles:    map[app.OperationType]string{},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationDeploy,
					PrincipalType: "component",
					PrincipalName: "test-component",
					Role:          "SpecificComponentRole",
				},
				{
					Operation:     app.OperationDeploy,
					PrincipalType: "component",
					PrincipalName: "*",
					Role:          "WildcardRole",
				},
			},
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "SpecificComponentRole",
			description:        "Specific matrix rule should take precedence over wildcard",
		},
		{
			name:              "component_role_missing_fallback_to_default",
			installDeployType: app.InstallDeployTypeApply,
			installDeployRole: "",
			componentRoles: map[app.OperationType]string{
				app.OperationDeploy: "MissingComponentRole", // This role won't be in stack outputs
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Should fallback to default when component role is missing from stack outputs",
		},
		{
			name:              "matrix_role_missing_fallback_to_default",
			installDeployType: app.InstallDeployTypeApply,
			installDeployRole: "",
			componentRoles:    map[app.OperationType]string{},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationDeploy,
					PrincipalType: "component",
					PrincipalName: "test-component",
					Role:          "MissingMatrixRole", // This role won't be in stack outputs
				},
			},
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Should fallback to default when matrix role is missing from stack outputs",
		},
		{
			name:               "runtime_role_missing_fallback_to_default",
			installDeployType:  app.InstallDeployTypeApply,
			installDeployRole:  "MissingRuntimeRole", // This role won't be in stack outputs
			componentRoles:     map[app.OperationType]string{},
			matrixRules:        nil,
			expectedOperation:  app.OperationDeploy,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Should fallback to default when runtime role is missing from stack outputs",
		},
		{
			name:              "matrix_role_missing_teardown_fallback",
			installDeployType: app.InstallDeployTypeTeardown,
			installDeployRole: "",
			componentRoles:    map[app.OperationType]string{},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationTeardown,
					PrincipalType: "component",
					PrincipalName: "test-component",
					Role:          "MissingTeardownRole", // This role won't be in stack outputs
				},
			},
			expectedOperation:  app.OperationTeardown,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "MaintenanceRole",
			description:        "Should fallback to default when teardown role is missing from stack outputs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appConfig := &app.AppConfig{
				OperationRoleConfig: app.AppOperationRoleConfig{
					Rules: tt.matrixRules,
				},
				PermissionsConfig: app.AppPermissionsConfig{
					CustomRoles: []app.AppAWSIAMRoleConfig{
						{Name: "ComponentDeployRole"},
						{Name: "ComponentTeardownRole"},
						{Name: "ComponentProvisionRole"},
						{Name: "MatrixDeployRole"},
						{Name: "MatrixTeardownRole"},
						{Name: "WildcardDeployRole"},
						{Name: "RuntimeDeployRole"},
						{Name: "RuntimeTeardownRole"},
						{Name: "RuntimeOnlyRole"},
						{Name: "SpecificComponentRole"},
						{Name: "WildcardRole"},
					},
					MaintenanceRole: app.AppAWSIAMRoleConfig{
						Name: "MaintenanceRole",
					},
					ProvisionRole: app.AppAWSIAMRoleConfig{
						Name: "ProvisionRole",
					},
					DeprovisionRole: app.AppAWSIAMRoleConfig{
						Name: "DeprovisionRole",
					},
				},
			}

			installDeploy := &app.InstallDeploy{
				Type: tt.installDeployType,
				Role: tt.installDeployRole,
			}

			opRole := make(pgtype.Hstore)
			for op, role := range tt.componentRoles {
				opRole[string(op)] = &role
			}
			build := &app.ComponentBuild{
				ComponentConfigConnection: app.ComponentConfigConnection{
					OperationRoles: opRole,
				},
			}

			comp := &app.Component{
				Name: "test-component",
			}

			stack := &app.InstallStack{
				InstallStackOutputs: app.InstallStackOutputs{
					AWSStackOutputs: &app.AWSStackOutputs{
						Region:                "us-west-2",
						ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/ProvisionRole",
						MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/MaintenanceRole",
						DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/DeprovisionRole",
						CustomRoleARNs: map[string]string{
							"MaintenanceRole":        "arn:aws:iam::123456789012:role/MaintenanceRole",
							"ProvisionRole":          "arn:aws:iam::123456789012:role/ProvisionRole",
							"DeprovisionRole":        "arn:aws:iam::123456789012:role/DeprovisionRole",
							"ComponentDeployRole":    "arn:aws:iam::123456789012:role/ComponentDeployRole",
							"ComponentTeardownRole":  "arn:aws:iam::123456789012:role/ComponentTeardownRole",
							"ComponentProvisionRole": "arn:aws:iam::123456789012:role/ComponentProvisionRole",
							"MatrixDeployRole":       "arn:aws:iam::123456789012:role/MatrixDeployRole",
							"MatrixTeardownRole":     "arn:aws:iam::123456789012:role/MatrixTeardownRole",
							"WildcardDeployRole":     "arn:aws:iam::123456789012:role/WildcardDeployRole",
							"RuntimeDeployRole":      "arn:aws:iam::123456789012:role/RuntimeDeployRole",
							"RuntimeTeardownRole":    "arn:aws:iam::123456789012:role/RuntimeTeardownRole",
							"RuntimeOnlyRole":        "arn:aws:iam::123456789012:role/RuntimeOnlyRole",
							"SpecificComponentRole":  "arn:aws:iam::123456789012:role/SpecificComponentRole",
							"WildcardRole":           "arn:aws:iam::123456789012:role/WildcardRole",
						},
					},
				},
			}

			// Create mock install state for testing
			installState := &state.State{
				ID:   "test-install",
				Name: "test-install",
			}

			roleSelection, operation, err := w.getRoleForDeploy(
				zap.NewNop(),
				appConfig,
				installDeploy,
				build,
				comp,
				stack,
				installState,
			)

			require.NoError(t, err, "getRoleForDeploy should not return error for test: %s", tt.description)
			assert.Equal(t, tt.expectedOperation, operation, "Operation type mismatch: %s", tt.description)
			assert.Equal(t, tt.expectedRoleSource, roleSelection.Source, "Role source mismatch: %s", tt.description)
			assert.Equal(t, tt.expectedRoleName, roleSelection.RoleName, "Role name mismatch: %s", tt.description)
			assert.NotEmpty(t, roleSelection.RoleARN, "Role ARN should be populated: %s", tt.description)
			assert.Contains(t, roleSelection.RoleARN, tt.expectedRoleName, "Role ARN should contain role name: %s", tt.description)

			t.Logf("%s: Got role=%s (source=%s, operation=%s)",
				tt.name,
				roleSelection.RoleName,
				roleSelection.Source,
				operation,
			)
		})
	}
}
