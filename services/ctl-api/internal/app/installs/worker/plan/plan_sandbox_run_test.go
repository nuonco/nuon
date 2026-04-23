package plan

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func TestGetPoliciesStableKeys(t *testing.T) {
	p := &Planner{}

	policies := []app.AppPolicyConfig{
		{
			Type: config.AppPolicyTypeKubernetesCluster,
			Contents: `kind: ClusterRole
metadata:
  name: kyverno:linkerd:manage
`,
		},
		{
			Type: config.AppPolicyTypeKubernetesCluster,
			Contents: `kind: ClusterPolicy
metadata:
  name: linkerd-authz-for-restate
`,
		},
	}

	// Run twice with the slice reordered — the resulting map must be identical.
	first, err := p.getPolicies(&app.AppPoliciesConfig{Policies: policies})
	require.NoError(t, err)

	reversed := []app.AppPolicyConfig{policies[1], policies[0]}
	second, err := p.getPolicies(&app.AppPoliciesConfig{Policies: reversed})
	require.NoError(t, err)

	assert.Equal(t, first, second, "reordering policies must not change the rendered key→content map")
	assert.Contains(t, first, "clusterrole-kyverno-linkerd-manage.yaml")
	assert.Contains(t, first, "clusterpolicy-linkerd-authz-for-restate.yaml")
}

func TestGetPoliciesDuplicateKeyErrors(t *testing.T) {
	p := &Planner{}

	policies := []app.AppPolicyConfig{
		{
			Type: config.AppPolicyTypeKubernetesCluster,
			Contents: `kind: ClusterRole
metadata:
  name: dup
`,
		},
		{
			Type: config.AppPolicyTypeKubernetesCluster,
			Contents: `kind: ClusterRole
metadata:
  name: dup
`,
		},
	}

	_, err := p.getPolicies(&app.AppPoliciesConfig{Policies: policies})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate policy manifest key")
}

func TestGetRoleForSandbox(t *testing.T) {
	p := &Planner{}
	l := zap.NewNop()

	tests := []struct {
		name               string
		sandboxRunType     app.SandboxRunType
		sandboxRuntimeRole string // Runtime role from sandboxRun.Role
		sandboxEntityRoles pgtype.Hstore
		matrixRules        []*app.AppOperationRoleRule
		expectedOperation  app.OperationType
		expectedRoleSource operationroles.RoleSelectionSource
		expectedRoleName   string
		description        string
	}{
		{
			name:               "no_rules_provision",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules:        nil,
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "ProvisionRole",
			description:        "Provision with no rules should use default provision role",
		},
		{
			name:               "no_rules_reprovision",
			sandboxRunType:     app.SandboxRunTypeReprovision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules:        nil,
			expectedOperation:  app.OperationReprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "ProvisionRole",
			description:        "Reprovision with no rules should use default provision role",
		},
		{
			name:               "no_rules_deprovision",
			sandboxRunType:     app.SandboxRunTypeDeprovision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules:        nil,
			expectedOperation:  app.OperationDeprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "DeprovisionRole",
			description:        "Deprovision with no rules should use default deprovision role",
		},
		{
			name:               "no_rules_unknown_type_defaults_to_provision",
			sandboxRunType:     app.SandboxRunType("unknown"),
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules:        nil,
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "ProvisionRole",
			description:        "Unknown run type should default to provision operation",
		},

		{
			name:               "entity_role_provision",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{
				"provision": generics.ToPtr("SandboxProvisionRole"),
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "SandboxProvisionRole",
			description:        "Provision should use sandbox-specific provision role",
		},
		{
			name:               "entity_role_reprovision",
			sandboxRunType:     app.SandboxRunTypeReprovision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{
				"reprovision": generics.ToPtr("SandboxReprovisionRole"),
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationReprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "SandboxReprovisionRole",
			description:        "Reprovision should use sandbox-specific reprovision role",
		},
		{
			name:               "entity_role_deprovision",
			sandboxRunType:     app.SandboxRunTypeDeprovision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{
				"deprovision": generics.ToPtr("SandboxDeprovisionRole"),
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationDeprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "SandboxDeprovisionRole",
			description:        "Deprovision should use sandbox-specific deprovision role",
		},
		{
			name:               "entity_has_different_operation_role",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{
				"deprovision": generics.ToPtr("SandboxDeprovisionRole"), // Different operation
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "ProvisionRole",
			description:        "Provision should fall back to default when sandbox only has deprovision role",
		},
		{
			name:               "matrix_rule_provision",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationProvision,
					PrincipalType: "sandbox",
					PrincipalName: "", // Sandboxes don't have names
					Role:          "MatrixProvisionRole",
				},
			},
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "MatrixProvisionRole",
			description:        "Provision should use matrix rule for sandbox",
		},
		{
			name:               "matrix_rule_reprovision",
			sandboxRunType:     app.SandboxRunTypeReprovision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationReprovision,
					PrincipalType: "sandbox",
					PrincipalName: "",
					Role:          "MatrixReprovisionRole",
				},
			},
			expectedOperation:  app.OperationReprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "MatrixReprovisionRole",
			description:        "Reprovision should use matrix rule for sandbox",
		},
		{
			name:               "matrix_rule_deprovision",
			sandboxRunType:     app.SandboxRunTypeDeprovision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationDeprovision,
					PrincipalType: "sandbox",
					PrincipalName: "",
					Role:          "MatrixDeprovisionRole",
				},
			},
			expectedOperation:  app.OperationDeprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "MatrixDeprovisionRole",
			description:        "Deprovision should use matrix rule for sandbox",
		},
		{
			name:               "matrix_rule_wrong_principal_type",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationProvision,
					PrincipalType: "component", // Wrong principal type
					PrincipalName: "*",
					Role:          "ComponentRole",
				},
			},
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "ProvisionRole",
			description:        "Sandbox provision should fall back to default when matrix rule is for components",
		},
		{
			name:               "matrix_rule_wrong_operation",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationDeprovision, // Wrong operation
					PrincipalType: "sandbox",
					PrincipalName: "",
					Role:          "MatrixDeprovisionRole",
				},
			},
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "ProvisionRole",
			description:        "Provision should fall back to default when matrix rule is for different operation",
		},

		{
			name:               "entity_role_overrides_matrix",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{
				"provision": generics.ToPtr("SandboxProvisionRole"),
			},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationProvision,
					PrincipalType: "sandbox",
					PrincipalName: "",
					Role:          "MatrixProvisionRole",
				},
			},
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "SandboxProvisionRole",
			description:        "Entity role should take precedence over matrix rule (entity > matrix)",
		},
		{
			name:               "both_present_deprovision",
			sandboxRunType:     app.SandboxRunTypeDeprovision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{
				"deprovision": generics.ToPtr("SandboxDeprovisionRole"),
			},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationDeprovision,
					PrincipalType: "sandbox",
					PrincipalName: "",
					Role:          "MatrixDeprovisionRole",
				},
			},
			expectedOperation:  app.OperationDeprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "SandboxDeprovisionRole",
			description:        "Entity deprovision role should override matrix rule",
		},

		{
			name:               "runtime_role_overrides_all_provision",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "RuntimeProvisionRole",
			sandboxEntityRoles: pgtype.Hstore{
				"provision": generics.ToPtr("SandboxProvisionRole"),
			},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationProvision,
					PrincipalType: "sandbox",
					PrincipalName: "",
					Role:          "MatrixProvisionRole",
				},
			},
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceRuntime,
			expectedRoleName:   "RuntimeProvisionRole",
			description:        "Runtime role should override all other sources (runtime > entity > matrix > default)",
		},
		{
			name:               "runtime_role_overrides_all_deprovision",
			sandboxRunType:     app.SandboxRunTypeDeprovision,
			sandboxRuntimeRole: "RuntimeDeprovisionRole",
			sandboxEntityRoles: pgtype.Hstore{
				"deprovision": generics.ToPtr("SandboxDeprovisionRole"),
			},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationDeprovision,
					PrincipalType: "sandbox",
					PrincipalName: "",
					Role:          "MatrixDeprovisionRole",
				},
			},
			expectedOperation:  app.OperationDeprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceRuntime,
			expectedRoleName:   "RuntimeDeprovisionRole",
			description:        "Runtime role should override all sources for deprovision",
		},
		{
			name:               "runtime_role_only",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "RuntimeOnlyRole",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules:        nil,
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceRuntime,
			expectedRoleName:   "RuntimeOnlyRole",
			description:        "Runtime role should work even when no other rules exist",
		},
		{
			name:               "runtime_role_reprovision",
			sandboxRunType:     app.SandboxRunTypeReprovision,
			sandboxRuntimeRole: "RuntimeReprovisionRole",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules:        nil,
			expectedOperation:  app.OperationReprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceRuntime,
			expectedRoleName:   "RuntimeReprovisionRole",
			description:        "Runtime role should work for reprovision operation",
		},

		{
			name:               "multiple_matrix_rules_first_match_wins",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationProvision,
					PrincipalType: "sandbox",
					PrincipalName: "",
					Role:          "FirstMatchRole",
				},
				{
					Operation:     app.OperationProvision,
					PrincipalType: "sandbox",
					PrincipalName: "",
					Role:          "SecondMatchRole",
				},
			},
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceMatrix,
			expectedRoleName:   "FirstMatchRole",
			description:        "First matching matrix rule should be used",
		},

		{
			name:               "empty_runtime_role_ignored",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{
				"provision": generics.ToPtr("SandboxProvisionRole"),
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "SandboxProvisionRole",
			description:        "Empty runtime role should not override entity role",
		},
		{
			name:               "all_operations_present_in_entity",
			sandboxRunType:     app.SandboxRunTypeDeprovision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{
				"provision":   generics.ToPtr("SandboxProvisionRole"),
				"reprovision": generics.ToPtr("SandboxReprovisionRole"),
				"deprovision": generics.ToPtr("SandboxDeprovisionRole"),
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationDeprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceEntity,
			expectedRoleName:   "SandboxDeprovisionRole",
			description:        "Should select correct operation role when all are defined",
		},

		{
			name:               "entity_role_missing_fallback_provision",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{
				"provision": generics.ToPtr("MissingEntityRole"), // Not in stack outputs
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "ProvisionRole",
			description:        "Should fallback to default provision role when entity role is missing",
		},
		{
			name:               "entity_role_missing_fallback_deprovision",
			sandboxRunType:     app.SandboxRunTypeDeprovision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{
				"deprovision": generics.ToPtr("MissingDeprovisionRole"), // Not in stack outputs
			},
			matrixRules:        nil,
			expectedOperation:  app.OperationDeprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "DeprovisionRole",
			description:        "Should fallback to default deprovision role when entity role is missing",
		},
		{
			name:               "matrix_role_missing_fallback_provision",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "",
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules: []*app.AppOperationRoleRule{
				{
					Operation:     app.OperationProvision,
					PrincipalType: "sandbox",
					PrincipalName: "",
					Role:          "MissingMatrixRole", // Not in stack outputs
				},
			},
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "ProvisionRole",
			description:        "Should fallback to default when matrix role is missing from stack outputs",
		},
		{
			name:               "runtime_role_missing_fallback_provision",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "MissingRuntimeRole", // Not in stack outputs
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules:        nil,
			expectedOperation:  app.OperationProvision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "ProvisionRole",
			description:        "Should fallback to default when runtime role is missing from stack outputs",
		},
		{
			name:               "runtime_role_missing_fallback_deprovision",
			sandboxRunType:     app.SandboxRunTypeDeprovision,
			sandboxRuntimeRole: "MissingRuntimeRole", // Not in stack outputs
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules:        nil,
			expectedOperation:  app.OperationDeprovision,
			expectedRoleSource: operationroles.RoleSelectionSourceDefault,
			expectedRoleName:   "DeprovisionRole",
			description:        "Should fallback to default deprovision when runtime role is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appConfig := &app.AppConfig{
				SandboxConfig: app.AppSandboxConfig{
					OperationRoles: tt.sandboxEntityRoles,
				},
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
						{Name: "SandboxProvisionRole"},
						{Name: "SandboxReprovisionRole"},
						{Name: "SandboxDeprovisionRole"},
						{Name: "MatrixProvisionRole"},
						{Name: "MatrixReprovisionRole"},
						{Name: "MatrixDeprovisionRole"},
						{Name: "RuntimeProvisionRole"},
						{Name: "RuntimeDeprovisionRole"},
						{Name: "RuntimeReprovisionRole"},
						{Name: "RuntimeOnlyRole"},
						{Name: "FirstMatchRole"},
						{Name: "SecondMatchRole"},
						{Name: "ComponentRole"},
					},
				},
				OperationRoleConfig: app.AppOperationRoleConfig{
					Rules: tt.matrixRules,
				},
			}

			sandboxRun := &app.InstallSandboxRun{
				RunType: tt.sandboxRunType,
				Role:    tt.sandboxRuntimeRole,
			}

			stack := &app.InstallStack{
				InstallStackOutputs: app.InstallStackOutputs{
					AWSStackOutputs: &app.AWSStackOutputs{
						Region:                "us-west-2",
						ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/ProvisionRole",
						MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/MaintenanceRole",
						DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/DeprovisionRole",
						CustomRoleARNs: map[string]string{
							"SandboxProvisionRole":   "arn:aws:iam::123456789012:role/SandboxProvisionRole",
							"SandboxReprovisionRole": "arn:aws:iam::123456789012:role/SandboxReprovisionRole",
							"SandboxDeprovisionRole": "arn:aws:iam::123456789012:role/SandboxDeprovisionRole",
							"MatrixProvisionRole":    "arn:aws:iam::123456789012:role/MatrixProvisionRole",
							"MatrixReprovisionRole":  "arn:aws:iam::123456789012:role/MatrixReprovisionRole",
							"MatrixDeprovisionRole":  "arn:aws:iam::123456789012:role/MatrixDeprovisionRole",
							"RuntimeProvisionRole":   "arn:aws:iam::123456789012:role/RuntimeProvisionRole",
							"RuntimeDeprovisionRole": "arn:aws:iam::123456789012:role/RuntimeDeprovisionRole",
							"RuntimeReprovisionRole": "arn:aws:iam::123456789012:role/RuntimeReprovisionRole",
							"RuntimeOnlyRole":        "arn:aws:iam::123456789012:role/RuntimeOnlyRole",
							"FirstMatchRole":         "arn:aws:iam::123456789012:role/FirstMatchRole",
							"SecondMatchRole":        "arn:aws:iam::123456789012:role/SecondMatchRole",
							"ComponentRole":          "arn:aws:iam::123456789012:role/ComponentRole",
						},
					},
				},
			}

			installState := &state.State{
				ID:   "test-install",
				Name: "test-install",
			}

			roleSelection, operation, err := p.getRoleForSandbox(
				l,
				appConfig,
				sandboxRun,
				stack,
				installState,
			)

			require.NoError(t, err, "getRoleForSandbox should not return error for test: %s", tt.description)
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
