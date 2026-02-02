package operationroles

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestResolveRoleARN(t *testing.T) {
	tests := []struct {
		name          string
		roleName      string
		appCfg        *app.AppConfig
		stackOutputs  *app.InstallStackOutputs
		expectedARN   string
		expectError   bool
		errorContains string
	}{
		{
			name:          "nil stack outputs",
			roleName:      "some-role",
			appCfg:        &app.AppConfig{},
			stackOutputs:  nil,
			expectedARN:   "",
			expectError:   true,
			errorContains: "no AWS stack outputs available",
		},
		{
			name:     "nil AWS stack outputs",
			roleName: "some-role",
			appCfg:   &app.AppConfig{},
			stackOutputs: &app.InstallStackOutputs{
				AWSStackOutputs: nil,
			},
			expectedARN:   "",
			expectError:   true,
			errorContains: "no AWS stack outputs available",
		},
		{
			name:     "role found in provision role",
			roleName: "provision",
			appCfg: &app.AppConfig{
				PermissionsConfig: app.AppPermissionsConfig{
					ProvisionRole: app.AppAWSIAMRoleConfig{
						Name: "provision",
					},
					MaintenanceRole: app.AppAWSIAMRoleConfig{
						Name: "maintenance",
					},
					DeprovisionRole: app.AppAWSIAMRoleConfig{
						Name: "deprovision",
					},
				},
			},
			stackOutputs: &app.InstallStackOutputs{
				AWSStackOutputs: &app.AWSStackOutputs{
					ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/provision-role",
					MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/maintenance-role",
					DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/deprovision-role",
					CustomRoleARNs:        make(map[string]string),
					BreakGlassRoleARNs:    make(map[string]string),
				},
			},
			expectedARN: "arn:aws:iam::123456789012:role/provision-role",
			expectError: false,
		},
		{
			name:     "role found in maintenance role",
			roleName: "maintenance",
			appCfg: &app.AppConfig{
				PermissionsConfig: app.AppPermissionsConfig{
					ProvisionRole: app.AppAWSIAMRoleConfig{
						Name: "provision",
					},
					MaintenanceRole: app.AppAWSIAMRoleConfig{
						Name: "maintenance",
					},
					DeprovisionRole: app.AppAWSIAMRoleConfig{
						Name: "deprovision",
					},
				},
			},
			stackOutputs: &app.InstallStackOutputs{
				AWSStackOutputs: &app.AWSStackOutputs{
					ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/provision-role",
					MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/maintenance-role",
					DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/deprovision-role",
					CustomRoleARNs:        make(map[string]string),
					BreakGlassRoleARNs:    make(map[string]string),
				},
			},
			expectedARN: "arn:aws:iam::123456789012:role/maintenance-role",
			expectError: false,
		},
		{
			name:     "role found in deprovision role",
			roleName: "deprovision",
			appCfg: &app.AppConfig{
				PermissionsConfig: app.AppPermissionsConfig{
					ProvisionRole: app.AppAWSIAMRoleConfig{
						Name: "provision",
					},
					MaintenanceRole: app.AppAWSIAMRoleConfig{
						Name: "maintenance",
					},
					DeprovisionRole: app.AppAWSIAMRoleConfig{
						Name: "deprovision",
					},
				},
			},
			stackOutputs: &app.InstallStackOutputs{
				AWSStackOutputs: &app.AWSStackOutputs{
					ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/provision-role",
					MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/maintenance-role",
					DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/deprovision-role",
					CustomRoleARNs:        make(map[string]string),
					BreakGlassRoleARNs:    make(map[string]string),
				},
			},
			expectedARN: "arn:aws:iam::123456789012:role/deprovision-role",
			expectError: false,
		},
		{
			name:     "role found in custom roles",
			roleName: "custom-db-role",
			appCfg: &app.AppConfig{
				PermissionsConfig: app.AppPermissionsConfig{
					ProvisionRole: app.AppAWSIAMRoleConfig{
						Name: "provision",
					},
					MaintenanceRole: app.AppAWSIAMRoleConfig{
						Name: "maintenance",
					},
					DeprovisionRole: app.AppAWSIAMRoleConfig{
						Name: "deprovision",
					},
					CustomRoles: []app.AppAWSIAMRoleConfig{
						{Name: "custom-db-role"},
						{Name: "custom-api-role"},
					},
				},
			},
			stackOutputs: &app.InstallStackOutputs{
				AWSStackOutputs: &app.AWSStackOutputs{
					ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/provision-role",
					MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/maintenance-role",
					DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/deprovision-role",
					CustomRoleARNs: map[string]string{
						"custom-db-role":  "arn:aws:iam::123456789012:role/custom-db",
						"custom-api-role": "arn:aws:iam::123456789012:role/custom-api",
					},
					BreakGlassRoleARNs: make(map[string]string),
				},
			},
			expectedARN: "arn:aws:iam::123456789012:role/custom-db",
			expectError: false,
		},
		{
			name:     "role found in break glass roles",
			roleName: "emergency-access",
			appCfg: &app.AppConfig{
				PermissionsConfig: app.AppPermissionsConfig{
					ProvisionRole: app.AppAWSIAMRoleConfig{
						Name: "provision",
					},
					MaintenanceRole: app.AppAWSIAMRoleConfig{
						Name: "maintenance",
					},
					DeprovisionRole: app.AppAWSIAMRoleConfig{
						Name: "deprovision",
					},
				},
				BreakGlassConfig: app.AppBreakGlassConfig{
					Roles: []app.AppAWSIAMRoleConfig{
						{Name: "emergency-access"},
						{Name: "db-migration-elevated"},
					},
				},
			},
			stackOutputs: &app.InstallStackOutputs{
				AWSStackOutputs: &app.AWSStackOutputs{
					ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/provision-role",
					MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/maintenance-role",
					DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/deprovision-role",
					CustomRoleARNs:        make(map[string]string),
					BreakGlassRoleARNs: map[string]string{
						"emergency-access":      "arn:aws:iam::123456789012:role/emergency",
						"db-migration-elevated": "arn:aws:iam::123456789012:role/db-migration",
					},
				},
			},
			expectedARN: "arn:aws:iam::123456789012:role/emergency",
			expectError: false,
		},
		{
			name:     "role not found in any category",
			roleName: "nonexistent-role",
			appCfg: &app.AppConfig{
				PermissionsConfig: app.AppPermissionsConfig{
					ProvisionRole: app.AppAWSIAMRoleConfig{
						Name: "provision",
					},
					MaintenanceRole: app.AppAWSIAMRoleConfig{
						Name: "maintenance",
					},
					DeprovisionRole: app.AppAWSIAMRoleConfig{
						Name: "deprovision",
					},
				},
			},
			stackOutputs: &app.InstallStackOutputs{
				AWSStackOutputs: &app.AWSStackOutputs{
					ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/provision-role",
					MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/maintenance-role",
					DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/deprovision-role",
					CustomRoleARNs:        make(map[string]string),
					BreakGlassRoleARNs:    make(map[string]string),
				},
			},
			expectedARN:   "",
			expectError:   true,
			errorContains: "role \"nonexistent-role\" not found in install stack outputs",
		},
		{
			name:     "custom role in config but not in stack outputs",
			roleName: "custom-missing-role",
			appCfg: &app.AppConfig{
				PermissionsConfig: app.AppPermissionsConfig{
					ProvisionRole: app.AppAWSIAMRoleConfig{
						Name: "provision",
					},
					MaintenanceRole: app.AppAWSIAMRoleConfig{
						Name: "maintenance",
					},
					DeprovisionRole: app.AppAWSIAMRoleConfig{
						Name: "deprovision",
					},
					CustomRoles: []app.AppAWSIAMRoleConfig{
						{Name: "custom-missing-role"},
					},
				},
			},
			stackOutputs: &app.InstallStackOutputs{
				AWSStackOutputs: &app.AWSStackOutputs{
					ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/provision-role",
					MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/maintenance-role",
					DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/deprovision-role",
					CustomRoleARNs:        make(map[string]string), // missing in stack outpout
					BreakGlassRoleARNs:    make(map[string]string),
				},
			},
			expectedARN:   "",
			expectError:   true,
			errorContains: "role \"custom-missing-role\" not found in install stack outputs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arn, err := resolveRoleARN(tt.roleName, tt.appCfg, tt.stackOutputs)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if arn != tt.expectedARN {
					t.Errorf("expected ARN %q, got %q", tt.expectedARN, arn)
				}
			}
		})
	}
}

func TestSelectRole(t *testing.T) {
	baseAppConfig := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			ProvisionRole: app.AppAWSIAMRoleConfig{
				Name: "provision",
			},
			MaintenanceRole: app.AppAWSIAMRoleConfig{
				Name: "maintenance",
			},
			DeprovisionRole: app.AppAWSIAMRoleConfig{
				Name: "deprovision",
			},
			CustomRoles: []app.AppAWSIAMRoleConfig{
				{Name: "custom-db-role"},
			},
		},
		BreakGlassConfig: app.AppBreakGlassConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{Name: "emergency-access"},
			},
		},
	}

	baseStackOutputs := &app.InstallStackOutputs{
		AWSStackOutputs: &app.AWSStackOutputs{
			ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/provision-role",
			MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/maintenance-role",
			DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/deprovision-role",
			CustomRoleARNs: map[string]string{
				"custom-db-role": "arn:aws:iam::123456789012:role/custom-db",
			},
			BreakGlassRoleARNs: map[string]string{
				"emergency-access": "arn:aws:iam::123456789012:role/emergency",
			},
		},
	}

	tests := []struct {
		name             string
		ctx              *SelectionContext
		expectedRoleName string
		expectedRoleARN  string
		expectedSource   RoleSelectionSource
		expectError      bool
		errorContains    string
	}{
		{
			name:          "nil context",
			ctx:           nil,
			expectError:   true,
			errorContains: "selection context is required",
		},
		{
			name: "runtime role takes precedence",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "emergency-access",
				EntityRoles: EntityOperationRoleMap{
					app.OperationDeploy: "custom-db-role",
				},
				MatrixRules: []*app.OperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "maintenance"},
				},
				DefaultRole:  "provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
			},
			expectedRoleName: "emergency-access",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/emergency",
			expectedSource:   RoleSelectionSourceRuntime,
			expectError:      false,
		},
		{
			name: "entity role takes precedence over matrix and default",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "",
				EntityRoles: EntityOperationRoleMap{
					app.OperationDeploy: "custom-db-role",
				},
				MatrixRules: []*app.OperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "maintenance"},
				},
				DefaultRole:  "provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
			},
			expectedRoleName: "custom-db-role",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/custom-db",
			expectedSource:   RoleSelectionSourceEntity,
			expectError:      false,
		},
		{
			name: "matrix rule takes precedence over default",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules: []*app.OperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "maintenance"},
				},
				DefaultRole:  "provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
			},
			expectedRoleName: "maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/maintenance-role",
			expectedSource:   RoleSelectionSourceMatrix,
			expectError:      false,
		},
		{
			name: "default role used when no other rules match",
			ctx: &SelectionContext{
				Operation:     app.OperationProvision,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "api",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "provision",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
			},
			expectedRoleName: "provision",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/provision-role",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "default role for sandbox provision",
			ctx: &SelectionContext{
				Operation:     app.OperationProvision,
				PrincipalType: principal.TypeSandbox,
				PrincipalName: "",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "provision",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
			},
			expectedRoleName: "provision",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/provision-role",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "default role for sandbox deprovision",
			ctx: &SelectionContext{
				Operation:     app.OperationDeprovision,
				PrincipalType: principal.TypeSandbox,
				PrincipalName: "",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "deprovision",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
			},
			expectedRoleName: "deprovision",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/deprovision-role",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "default role for action trigger",
			ctx: &SelectionContext{
				Operation:     app.OperationTrigger,
				PrincipalType: principal.TypeAction,
				PrincipalName: "deploy-action",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "maintenance",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
			},
			expectedRoleName: "maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/maintenance-role",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "error when default role not found",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "api",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "nonexistent-role",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
			},
			expectError:   true,
			errorContains: "default role \"nonexistent-role\"",
		},
		{
			name: "error when runtime role not found",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "invalid-role",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "provision",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
			},
			expectError:   true,
			errorContains: "runtime role \"invalid-role\"",
		},
		{
			name: "wildcard matrix rule matches component",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "any-component",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules: []*app.OperationRoleRule{
					{PrincipalType: "component", PrincipalName: "*", Operation: app.OperationDeploy, Role: "maintenance"},
				},
				DefaultRole:  "provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
			},
			expectedRoleName: "maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/maintenance-role",
			expectedSource:   RoleSelectionSourceMatrix,
			expectError:      false,
		},
		{
			name: "wildcard matrix rule matches action",
			ctx: &SelectionContext{
				Operation:     app.OperationTrigger,
				PrincipalType: principal.TypeAction,
				PrincipalName: "any-action",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules: []*app.OperationRoleRule{
					{PrincipalType: "action", PrincipalName: "*", Operation: app.OperationTrigger, Role: "emergency-access"},
				},
				DefaultRole:  "maintenance",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
			},
			expectedRoleName: "emergency-access",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/emergency",
			expectedSource:   RoleSelectionSourceMatrix,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SelectRole(tt.ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result.RoleName != tt.expectedRoleName {
					t.Errorf("expected role name %q, got %q", tt.expectedRoleName, result.RoleName)
				}
				if result.RoleARN != tt.expectedRoleARN {
					t.Errorf("expected role ARN %q, got %q", tt.expectedRoleARN, result.RoleARN)
				}
				if result.Source != tt.expectedSource {
					t.Errorf("expected source %q, got %q", tt.expectedSource, result.Source)
				}
			}
		})
	}
}
