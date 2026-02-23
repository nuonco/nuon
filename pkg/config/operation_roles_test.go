package config

import (
	"testing"
)

func TestOperationRoleRule_ValidatePrincipal(t *testing.T) {
	tests := []struct {
		name          string
		principal     string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid component with name",
			principal:   "nuon::component:database",
			expectError: false,
		},
		{
			name:        "valid component with wildcard",
			principal:   "nuon::component:*",
			expectError: false,
		},
		{
			name:        "valid sandbox",
			principal:   "nuon::sandbox",
			expectError: false,
		},
		{
			name:        "valid action with name",
			principal:   "nuon::action:migrate",
			expectError: false,
		},
		{
			name:          "sandbox with name should fail",
			principal:     "nuon::sandbox:test",
			expectError:   true,
			errorContains: "sandbox principal should not have a name",
		},
		{
			name:          "component without name should fail",
			principal:     "nuon::component",
			expectError:   true,
			errorContains: "component principal must have a name",
		},
		{
			name:          "action without name should fail",
			principal:     "nuon::action",
			expectError:   true,
			errorContains: "action principal must have a name",
		},
		{
			name:          "empty principal",
			principal:     "",
			expectError:   true,
			errorContains: "principal cannot be empty",
		},
		{
			name:          "invalid principal type",
			principal:     "nuon::invalid:name",
			expectError:   true,
			errorContains: "principal type must be one of",
		},
		{
			name:          "missing nuon prefix",
			principal:     "component:database",
			expectError:   true,
			errorContains: "must start with 'nuon::'",
		},
		{
			name:          "component with empty name after colon",
			principal:     "nuon::component:",
			expectError:   true,
			errorContains: "component principal must have a name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &OperationRoleRule{
				Principal: tt.principal,
			}

			err := rule.ValidatePrincipal()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestOperationRoleRule_Validate(t *testing.T) {
	tests := []struct {
		name          string
		rule          OperationRoleRule
		expectError   bool
		errorContains string
	}{
		{
			name: "valid rule with component",
			rule: OperationRoleRule{
				Principal: "nuon::component:database",
				Operation: OperationProvision,
				RoleName:  "db_provision_role",
			},
			expectError: false,
		},
		{
			name: "valid rule with sandbox",
			rule: OperationRoleRule{
				Principal: "nuon::sandbox",
				Operation: OperationDeprovision,
				RoleName:  "sandbox_cleanup",
			},
			expectError: false,
		},
		{
			name: "valid rule with action",
			rule: OperationRoleRule{
				Principal: "nuon::action:migrate",
				Operation: OperationTrigger,
				RoleName:  "migration_role",
			},
			expectError: false,
		},
		{
			name: "empty role name",
			rule: OperationRoleRule{
				Principal: "nuon::component:database",
				Operation: OperationProvision,
				RoleName:  "",
			},
			expectError:   true,
			errorContains: "role name cannot be empty",
		},
		{
			name: "whitespace-only role name",
			rule: OperationRoleRule{
				Principal: "nuon::component:database",
				Operation: OperationProvision,
				RoleName:  "   ",
			},
			expectError:   true,
			errorContains: "role name cannot be empty",
		},
		{
			name: "invalid principal",
			rule: OperationRoleRule{
				Principal: "invalid::component:db",
				Operation: OperationProvision,
				RoleName:  "db_role",
			},
			expectError:   true,
			errorContains: "must start with 'nuon::'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestOperationRolesConfig_ValidateWithConfig(t *testing.T) {
	// Setup test components and actions
	components := []*Component{
		{Name: "database"},
		{Name: "api-server"},
		{Name: "worker"},
	}

	actions := []*ActionConfig{
		{Name: "migrate"},
		{Name: "seed-data"},
		{Name: "post-deploy-check"},
	}

	// Setup test permissions config
	permissions := &PermissionsConfig{
		ProvisionRole:   &AppAWSIAMRole{Name: "custom_provision"},
		MaintenanceRole: &AppAWSIAMRole{Name: "custom_maintenance"},
		DeprovisionRole: &AppAWSIAMRole{Name: "custom_deprovision"},
		Roles: []*AppAWSIAMRole{
			{Name: "db_role"},
			{Name: "migration_role"},
			{Name: "sandbox_role"},
			{Name: "all_components_role"},
			{Name: "all_actions_role"},
			{Name: "db_provision"},
			{Name: "api_update"},
			{Name: "migrate_role"},
			{Name: "some_role"},
			{Name: "custom_provision"},
			{Name: "custom_maintenance"},
			{Name: "custom_deprovision"},
			{Name: "{{ .nuon.install.id }}-custom_deprovision"},
		},
	}

	// Setup test break glass config
	breakGlass := &BreakGlass{
		Roles: []*AppAWSIAMRole{
			{Name: "emergency_access"},
			{Name: "db_migration_elevated"},
		},
	}

	tests := []struct {
		name          string
		config        *OperationRolesConfig
		components    []*Component
		actions       []*ActionConfig
		permissions   *PermissionsConfig
		breakGlass    *BreakGlass
		expectError   bool
		errorContains string
	}{
		{
			name: "valid config with existing component and role in permissions",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::component:database",
						Operation: OperationProvision,
						RoleName:  "db_role",
					},
				},
			},
			components:  components,
			actions:     actions,
			permissions: permissions,
			breakGlass:  breakGlass,
			expectError: false,
		},
		{
			name: "valid config with rendered role without spaces",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::component:database",
						Operation: OperationProvision,
						RoleName:  "{{.nuon.install.id}}-custom_deprovision",
					},
				},
			},
			components:  components,
			actions:     actions,
			permissions: permissions,
			breakGlass:  breakGlass,
			expectError: false,
		},
		{
			name: "valid config with rendered role with spaces",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::component:database",
						Operation: OperationProvision,
						RoleName:  "{{ .nuon.install.id }}-custom_deprovision",
					},
				},
			},
			components:  components,
			actions:     actions,
			permissions: permissions,
			breakGlass:  breakGlass,
			expectError: false,
		},
		{
			name: "valid config with existing action",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::action:migrate",
						Operation: OperationTrigger,
						RoleName:  "migration_role",
					},
				},
			},
			components:  components,
			actions:     actions,
			permissions: permissions,
			breakGlass:  breakGlass,
			expectError: false,
		},
		{
			name: "valid config with sandbox",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::sandbox",
						Operation: OperationDeprovision,
						RoleName:  "sandbox_role",
					},
				},
			},
			components:  components,
			actions:     actions,
			permissions: permissions,
			breakGlass:  breakGlass,
			expectError: false,
		},
		{
			name: "valid config with multiple rules",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::component:database",
						Operation: OperationProvision,
						RoleName:  "db_provision",
					},
					{
						Principal: "nuon::component:api-server",
						Operation: OperationDeploy,
						RoleName:  "api_update",
					},
					{
						Principal: "nuon::action:migrate",
						Operation: OperationTrigger,
						RoleName:  "migrate_role",
					},
				},
			},
			components:  components,
			actions:     actions,
			permissions: permissions,
			breakGlass:  breakGlass,
			expectError: false,
		},
		{
			name: "component does not exist",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::component:nonexistent",
						Operation: OperationProvision,
						RoleName:  "some_role",
					},
				},
			},
			components:    components,
			actions:       actions,
			permissions:   permissions,
			breakGlass:    breakGlass,
			expectError:   true,
			errorContains: "component \"nonexistent\" does not exist in app config",
		},
		{
			name: "action does not exist",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::action:nonexistent-action",
						Operation: OperationTrigger,
						RoleName:  "some_role",
					},
				},
			},
			components:    components,
			actions:       actions,
			permissions:   permissions,
			breakGlass:    breakGlass,
			expectError:   true,
			errorContains: "action \"nonexistent-action\" does not exist in app config",
		},
		{
			name: "empty components list with component rule",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::component:database",
						Operation: OperationProvision,
						RoleName:  "db_role",
					},
				},
			},
			components:    []*Component{},
			actions:       actions,
			permissions:   permissions,
			breakGlass:    breakGlass,
			expectError:   true,
			errorContains: "component \"database\" does not exist in app config",
		},
		{
			name: "empty actions list with action rule",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::action:migrate",
						Operation: OperationTrigger,
						RoleName:  "migrate_role",
					},
				},
			},
			components:    components,
			actions:       []*ActionConfig{},
			permissions:   permissions,
			breakGlass:    breakGlass,
			expectError:   true,
			errorContains: "action \"migrate\" does not exist in app config",
		},
		{
			name: "multiple rules with one invalid",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::component:database",
						Operation: OperationProvision,
						RoleName:  "db_role",
					},
					{
						Principal: "nuon::component:invalid",
						Operation: OperationDeploy,
						RoleName:  "invalid_role",
					},
				},
			},
			components:    components,
			actions:       actions,
			permissions:   permissions,
			breakGlass:    breakGlass,
			expectError:   true,
			errorContains: "component \"invalid\" does not exist",
		},
		{
			name: "role does not exist in permissions or break glass",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::component:database",
						Operation: OperationProvision,
						RoleName:  "nonexistent_role",
					},
				},
			},
			components:    components,
			actions:       actions,
			permissions:   permissions,
			breakGlass:    breakGlass,
			expectError:   true,
			errorContains: "role \"nonexistent_role\" does not exist in permissions or break_glass config",
		},
		{
			name: "role exists in break glass config",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::action:migrate",
						Operation: OperationTrigger,
						RoleName:  "emergency_access",
					},
				},
			},
			components:  components,
			actions:     actions,
			permissions: permissions,
			breakGlass:  breakGlass,
			expectError: false,
		},
		{
			name: "role exists in standard roles",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::component:database",
						Operation: OperationProvision,
						RoleName:  "custom_provision",
					},
				},
			},
			components:  components,
			actions:     actions,
			permissions: permissions,
			breakGlass:  breakGlass,
			expectError: false,
		},
		{
			name:        "nil config",
			config:      nil,
			components:  components,
			actions:     actions,
			permissions: permissions,
			breakGlass:  breakGlass,
			expectError: false,
		},
		{
			name: "nil permissions config still validates standard roles",
			config: &OperationRolesConfig{
				Type: "matrix",
				RuleMatrix: []*OperationRoleRule{
					{
						Principal: "nuon::component:database",
						Operation: OperationProvision,
						RoleName:  "some_role",
					},
				},
			},
			components:    components,
			actions:       actions,
			permissions:   nil,
			breakGlass:    nil,
			expectError:   true,
			errorContains: "role \"some_role\" does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateWithConfig(tt.components, tt.actions, tt.permissions, tt.breakGlass)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
