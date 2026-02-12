package operationroles

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"go.uber.org/zap"
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
			errorContains: "unable to resolve runtime role arn \"invalid-role\"",
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
		// Azure Support Tests
		{
			name: "azure returns placeholder values",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "maintenance",
				AppConfig:     baseAppConfig,
				StackOutputs: &app.InstallStackOutputs{
					AzureStackOutputs: &app.AzureStackOutputs{
						SubscriptionID: "test-subscription-id",
					},
				},
			},
			expectedRoleName: "azure-placeholder-name",
			expectedRoleARN:  "azure-placeholder-arn",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "azure early exit ignores all other rules",
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
				DefaultRole: "maintenance",
				AppConfig:   baseAppConfig,
				StackOutputs: &app.InstallStackOutputs{
					AzureStackOutputs: &app.AzureStackOutputs{
						SubscriptionID: "test-subscription-id",
					},
				},
			},
			expectedRoleName: "azure-placeholder-name",
			expectedRoleARN:  "azure-placeholder-arn",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		// Sandbox Mode Tests
		{
			name: "sandbox mode returns default role when no runtime role",
			ctx: &SelectionContext{
				SandboxMode:   true,
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
			name: "sandbox mode respects runtime role",
			ctx: &SelectionContext{
				SandboxMode:   true,
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "emergency-access",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "maintenance",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
			},
			expectedRoleName: "emergency-access",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/emergency",
			expectedSource:   RoleSelectionSourceRuntime,
			expectError:      false,
		},
		{
			name: "sandbox mode ignores entity roles",
			ctx: &SelectionContext{
				SandboxMode:   true,
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "",
				EntityRoles: EntityOperationRoleMap{
					app.OperationDeploy: "custom-db-role",
				},
				MatrixRules:  nil,
				DefaultRole:  "maintenance",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
			},
			expectedRoleName: "maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/maintenance-role",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "sandbox mode ignores matrix rules",
			ctx: &SelectionContext{
				SandboxMode:   true,
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules: []*app.OperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "custom-db-role"},
				},
				DefaultRole:  "maintenance",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
			},
			expectedRoleName: "maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/maintenance-role",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "sandbox mode with runtime role overrides everything",
			ctx: &SelectionContext{
				SandboxMode:   true,
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "custom-db-role",
				EntityRoles: EntityOperationRoleMap{
					app.OperationDeploy: "emergency-access",
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
			expectedSource:   RoleSelectionSourceRuntime,
			expectError:      false,
		},
		{
			name: "sandbox mode ignores break glass when no runtime role",
			ctx: &SelectionContext{
				SandboxMode:    true,
				Operation:      app.OperationDeploy,
				PrincipalType:  principal.TypeComponent,
				PrincipalName:  "database",
				RuntimeRole:    "",
				BreakGlassRole: "emergency-access",
				EntityRoles:    nil,
				MatrixRules:    nil,
				DefaultRole:    "maintenance",
				AppConfig:      baseAppConfig,
				StackOutputs:   baseStackOutputs,
			},
			expectedRoleName: "maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/maintenance-role",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		// Break Glass Role Tests
		{
			name: "break glass role used when specified",
			ctx: &SelectionContext{
				Operation:      app.OperationTrigger,
				PrincipalType:  principal.TypeAction,
				PrincipalName:  "emergency-deploy",
				RuntimeRole:    "",
				BreakGlassRole: "emergency-access",
				EntityRoles:    nil,
				MatrixRules:    nil,
				DefaultRole:    "maintenance",
				AppConfig:      baseAppConfig,
				StackOutputs:   baseStackOutputs,
			},
			expectedRoleName: "emergency-access",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/emergency",
			expectedSource:   RoleSelectionSourceBreakGlass,
			expectError:      false,
		},
		{
			name: "break glass role takes precedence over entity role",
			ctx: &SelectionContext{
				Operation:      app.OperationDeploy,
				PrincipalType:  principal.TypeComponent,
				PrincipalName:  "database",
				RuntimeRole:    "",
				BreakGlassRole: "emergency-access",
				EntityRoles: EntityOperationRoleMap{
					app.OperationDeploy: "custom-db-role",
				},
				MatrixRules:  nil,
				DefaultRole:  "maintenance",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
			},
			expectedRoleName: "emergency-access",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/emergency",
			expectedSource:   RoleSelectionSourceBreakGlass,
			expectError:      false,
		},
		{
			name: "break glass role takes precedence over matrix rules",
			ctx: &SelectionContext{
				Operation:      app.OperationDeploy,
				PrincipalType:  principal.TypeComponent,
				PrincipalName:  "database",
				RuntimeRole:    "",
				BreakGlassRole: "emergency-access",
				EntityRoles:    nil,
				MatrixRules: []*app.OperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "maintenance"},
				},
				DefaultRole:  "provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
			},
			expectedRoleName: "emergency-access",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/emergency",
			expectedSource:   RoleSelectionSourceBreakGlass,
			expectError:      false,
		},
		{
			name: "break glass role takes precedence over default",
			ctx: &SelectionContext{
				Operation:      app.OperationDeploy,
				PrincipalType:  principal.TypeComponent,
				PrincipalName:  "database",
				RuntimeRole:    "",
				BreakGlassRole: "emergency-access",
				EntityRoles:    nil,
				MatrixRules:    nil,
				DefaultRole:    "maintenance",
				AppConfig:      baseAppConfig,
				StackOutputs:   baseStackOutputs,
			},
			expectedRoleName: "emergency-access",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/emergency",
			expectedSource:   RoleSelectionSourceBreakGlass,
			expectError:      false,
		},
		{
			name: "runtime role takes precedence over break glass",
			ctx: &SelectionContext{
				Operation:      app.OperationDeploy,
				PrincipalType:  principal.TypeComponent,
				PrincipalName:  "database",
				RuntimeRole:    "custom-db-role",
				BreakGlassRole: "emergency-access",
				EntityRoles:    nil,
				MatrixRules:    nil,
				DefaultRole:    "maintenance",
				AppConfig:      baseAppConfig,
				StackOutputs:   baseStackOutputs,
			},
			expectedRoleName: "custom-db-role",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/custom-db",
			expectedSource:   RoleSelectionSourceRuntime,
			expectError:      false,
		},
		{
			name: "error when break glass role not found",
			ctx: &SelectionContext{
				Operation:      app.OperationDeploy,
				PrincipalType:  principal.TypeComponent,
				PrincipalName:  "database",
				RuntimeRole:    "",
				BreakGlassRole: "nonexistent-break-glass",
				EntityRoles:    nil,
				MatrixRules:    nil,
				DefaultRole:    "maintenance",
				AppConfig:      baseAppConfig,
				StackOutputs:   baseStackOutputs,
			},
			expectError:   true,
			errorContains: "unable to resolve break glass role \"nonexistent-break-glass\"",
		},
		{
			name: "break glass role with all other rules present",
			ctx: &SelectionContext{
				Operation:      app.OperationDeploy,
				PrincipalType:  principal.TypeComponent,
				PrincipalName:  "database",
				RuntimeRole:    "",
				BreakGlassRole: "emergency-access",
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
			expectedSource:   RoleSelectionSourceBreakGlass,
			expectError:      false,
		},
		// Combined Scenarios
		{
			name: "azure with sandbox mode returns azure placeholder",
			ctx: &SelectionContext{
				SandboxMode:   true,
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "maintenance",
				AppConfig:     baseAppConfig,
				StackOutputs: &app.InstallStackOutputs{
					AzureStackOutputs: &app.AzureStackOutputs{
						SubscriptionID: "test-subscription-id",
					},
				},
			},
			expectedRoleName: "azure-placeholder-name",
			expectedRoleARN:  "azure-placeholder-arn",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "azure with break glass returns azure placeholder",
			ctx: &SelectionContext{
				Operation:      app.OperationDeploy,
				PrincipalType:  principal.TypeComponent,
				PrincipalName:  "database",
				RuntimeRole:    "",
				BreakGlassRole: "emergency-access",
				EntityRoles:    nil,
				MatrixRules:    nil,
				DefaultRole:    "maintenance",
				AppConfig:      baseAppConfig,
				StackOutputs: &app.InstallStackOutputs{
					AzureStackOutputs: &app.AzureStackOutputs{
						SubscriptionID: "test-subscription-id",
					},
				},
			},
			expectedRoleName: "azure-placeholder-name",
			expectedRoleARN:  "azure-placeholder-arn",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SelectRole(tt.ctx, zap.NewNop())

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

func TestRenderRoleName(t *testing.T) {
	tests := []struct {
		name          string
		roleName      string
		installState  *state.State
		expectedName  string
		expectError   bool
		errorContains string
	}{
		{
			name:     "role name with template resolves install name",
			roleName: "{{.nuon.install.name}}-role",
			installState: &state.State{
				Install: &state.InstallState{
					Name: "production",
				},
			},
			expectedName: "production-role",
			expectError:  false,
		},
		{
			name:     "role name without template returns unchanged",
			roleName: "maintenance-role",
			installState: &state.State{
				Install: &state.InstallState{
					Name: "production",
				},
			},
			expectedName: "maintenance-role",
			expectError:  false,
		},
		{
			name:         "nil install state returns role name unchanged",
			roleName:     "{{.nuon.install.name}}-role",
			installState: nil,
			expectedName: "{{.nuon.install.name}}-role",
			expectError:  false,
		},
		{
			name:     "empty role name returns empty",
			roleName: "",
			installState: &state.State{
				Install: &state.InstallState{
					Name: "production",
				},
			},
			expectedName: "",
			expectError:  false,
		},
		{
			name:     "invalid template syntax returns error",
			roleName: "{{.nuon.install.name",
			installState: &state.State{
				Install: &state.InstallState{
					Name: "production",
				},
			},
			expectError:   true,
			errorContains: "unable to render role name template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderRoleName(tt.roleName, tt.installState)

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
				if result != tt.expectedName {
					t.Errorf("expected %q, got %q", tt.expectedName, result)
				}
			}
		})
	}
}
