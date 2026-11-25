package plan

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
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
  name: linkerd-authz-for-payments
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
	assert.Contains(t, first, "clusterpolicy-linkerd-authz-for-payments.yaml")
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
		expectedError      string
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
			name:               "runtime_role_missing_provision_returns_error",
			sandboxRunType:     app.SandboxRunTypeProvision,
			sandboxRuntimeRole: "MissingRuntimeRole", // Not in stack outputs
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules:        nil,
			expectedError:      `unable to use requested role "MissingRuntimeRole"`,
			description:        "Should reject a runtime role missing from stack outputs",
		},
		{
			name:               "runtime_role_missing_deprovision_returns_error",
			sandboxRunType:     app.SandboxRunTypeDeprovision,
			sandboxRuntimeRole: "MissingRuntimeRole", // Not in stack outputs
			sandboxEntityRoles: pgtype.Hstore{},
			matrixRules:        nil,
			expectedError:      `unable to use requested role "MissingRuntimeRole"`,
			description:        "Should reject a runtime role missing from stack outputs",
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
				false,
				false,
			)

			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError)
				require.Nil(t, roleSelection)
				require.Empty(t, operation)
				return
			}
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

func TestGetSandboxRunTerraformVarsInstallTags(t *testing.T) {
	p := &Planner{}

	t.Run("aws sandboxes get both install tag keys", func(t *testing.T) {
		vars, err := p.getSandboxRunTerraformVars(&app.AppConfig{
			RunnerConfig: app.AppRunnerConfig{Type: app.AppRunnerTypeAWS},
		}, "nuon.co")
		require.NoError(t, err)

		tags, ok := vars["tags"].(map[string]string)
		require.True(t, ok, "tags var should be a string map, got %T", vars["tags"])

		// install locks filter the nuke on these
		assert.Equal(t, "{{.nuon.install.id}}", tags["install.nuon.co/id"])
		assert.Equal(t, "{{.nuon.install.id}}", tags["NUON_INSTALL_ID"])
	})

	t.Run("gcp sandboxes get the install label", func(t *testing.T) {
		vars, err := p.getSandboxRunTerraformVars(&app.AppConfig{
			RunnerConfig: app.AppRunnerConfig{Type: app.AppRunnerTypeGCP},
		}, "nuon.co")
		require.NoError(t, err)

		// gcp label keys can't contain dots or slashes
		labels, ok := vars["labels"].(map[string]string)
		require.True(t, ok, "labels var should be a string map, got %T", vars["labels"])
		assert.Equal(t, "{{.nuon.install.id}}", labels["nuon-install-id"])
	})
}

func TestRenderSandboxRunPlanPoliciesFromInstallACV(t *testing.T) {
	p := &Planner{}
	l := zap.NewNop()

	installACV := &app.AppConfig{
		ID: "acme-install-acv",
		SandboxConfig: app.AppSandboxConfig{
			Variables: pgtype.Hstore{"from": generics.ToPtr("install-acv")},
		},
		RunnerConfig: app.AppRunnerConfig{Type: app.AppRunnerTypeAWS},
		PoliciesConfig: app.AppPoliciesConfig{
			Policies: []app.AppPolicyConfig{
				{
					Type: config.AppPolicyTypeKubernetesCluster,
					Name: "install-acv-policy",
					Contents: `kind: ClusterPolicy
metadata:
  name: install-acv-policy
`,
				},
			},
		},
		PermissionsConfig: app.AppPermissionsConfig{
			ProvisionRole: app.AppAWSIAMRoleConfig{Name: "ProvisionRole"},
		},
	}
	targetACV := &app.AppConfig{
		ID: "acme-target-acv",
		SandboxConfig: app.AppSandboxConfig{
			Variables: pgtype.Hstore{"from": generics.ToPtr("target-acv")},
		},
		RunnerConfig: app.AppRunnerConfig{Type: app.AppRunnerTypeAWS},
		PoliciesConfig: app.AppPoliciesConfig{
			Policies: []app.AppPolicyConfig{
				{
					Type: config.AppPolicyTypeKubernetesCluster,
					Name: "target-acv-policy",
					Contents: `kind: ClusterPolicy
metadata:
  name: target-acv-policy
`,
				},
			},
		},
		PermissionsConfig: app.AppPermissionsConfig{
			ProvisionRole: app.AppAWSIAMRoleConfig{Name: "ProvisionRole"},
		},
	}

	newInput := func(policyCfg *app.AppConfig) *RenderSandboxRunPlanInput {
		return &RenderSandboxRunPlanInput{
			RootDomain: "example.com",
			Install:    &app.Install{ID: "acme-install", AppID: "acme-app"},
			Stack: &app.InstallStack{
				InstallStackOutputs: app.InstallStackOutputs{
					AWSStackOutputs: &app.AWSStackOutputs{
						Region:              "us-west-2",
						ProvisionIAMRoleARN: "arn:aws:iam::123456789012:role/ProvisionRole",
					},
				},
			},
			AppCfg:    targetACV,
			PolicyCfg: policyCfg,
			Run:       &app.InstallSandboxRun{RunType: app.SandboxRunTypeProvision},
			ResolveState: func() (*state.State, map[string]any, error) {
				return &state.State{}, map[string]any{
					"nuon": map[string]any{
						"install": map[string]any{"id": "acme-install"},
						"install_stack": map[string]any{
							"outputs": map[string]any{
								"region":                   "us-west-2",
								"vpc_id":                   "vpc-acme",
								"provision_iam_role_arn":   "arn:aws:iam::123456789012:role/ProvisionRole",
								"deprovision_iam_role_arn": "arn:aws:iam::123456789012:role/DeprovisionRole",
								"maintenance_iam_role_arn": "arn:aws:iam::123456789012:role/MaintenanceRole",
							},
						},
					},
				}, nil
			},
			ResolveSource: func() (*plantypes.GitSource, *plantypes.OCISource, error) {
				return nil, nil, nil
			},
		}
	}

	tests := map[string]struct {
		policyCfg         *app.AppConfig
		expectedPolicyKey string
		description       string
	}{
		"policy_cfg_set_renders_install_acv_policies": {
			policyCfg:         installACV,
			expectedPolicyKey: "clusterpolicy-install-acv-policy.yaml",
			description:       "the plan must embed the install's desired ACV policies, which is what the policy check evaluated",
		},
		"policy_cfg_nil_falls_back_to_target_acv": {
			policyCfg:         nil,
			expectedPolicyKey: "clusterpolicy-target-acv-policy.yaml",
			description:       "without a policy cfg the plan keeps rendering AppCfg policies (the other caller's path)",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			plan, _, err := p.RenderSandboxRunPlan(l, newInput(tt.policyCfg))
			require.NoError(t, err, tt.description)

			assert.Equal(t, "acme-target-acv", plan.AppConfigID, "the target ACV must still drive the plan identity")
			assert.Equal(t, "target-acv", plan.Vars["from"], "the target ACV must still drive the sandbox variables")
			assert.Len(t, plan.Policies, 1)
			assert.Contains(t, plan.Policies, tt.expectedPolicyKey, tt.description)
		})
	}
}
