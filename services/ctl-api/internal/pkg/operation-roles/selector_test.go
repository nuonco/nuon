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
	// Create a basic install state for tests
	installState := &state.State{
		Install: &state.InstallState{
			Name: "test-install",
		},
	}

	tests := []struct {
		name          string
		roleName      string
		appCfg        *app.AppConfig
		stackOutputs  *app.InstallStackOutputs
		installState  *state.State
		expectedARN   string
		expectError   bool
		errorContains string
	}{
		{
			name:          "nil stack outputs",
			roleName:      "some-role",
			appCfg:        &app.AppConfig{},
			stackOutputs:  nil,
			installState:  installState,
			expectedARN:   "",
			expectError:   true,
			errorContains: "stack outputs are required",
		},
		{
			name:     "nil AWS and Azure stack outputs",
			roleName: "some-role",
			appCfg:   &app.AppConfig{},
			stackOutputs: &app.InstallStackOutputs{
				AWSStackOutputs:   nil,
				AzureStackOutputs: nil,
			},
			installState:  installState,
			expectedARN:   "",
			expectError:   true,
			errorContains: "stack outputs must have either AWS, Azure, or GCP outputs",
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
			installState: installState,
			expectedARN:  "arn:aws:iam::123456789012:role/provision-role",
			expectError:  false,
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
			installState: installState,
			expectedARN:  "arn:aws:iam::123456789012:role/maintenance-role",
			expectError:  false,
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
			installState: installState,
			expectedARN:  "arn:aws:iam::123456789012:role/deprovision-role",
			expectError:  false,
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
			installState: installState,
			expectedARN:  "arn:aws:iam::123456789012:role/custom-db",
			expectError:  false,
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
			installState: installState,
			expectedARN:  "arn:aws:iam::123456789012:role/emergency",
			expectError:  false,
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
			installState:  installState,
			expectedARN:   "",
			expectError:   true,
			errorContains: "role nonexistent-role not found in install stack outputs",
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
			installState:  installState,
			expectedARN:   "",
			expectError:   true,
			errorContains: "role custom-missing-role not found in install stack outputs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arn, err := resolveRoleARN(tt.roleName, tt.appCfg, tt.stackOutputs, tt.installState)

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
	// Create a basic install state for tests
	baseInstallState := &state.State{
		Install: &state.InstallState{
			Name: "test-install",
			ID:   "1234",
		},
	}

	baseAppConfig := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			ProvisionRole: app.AppAWSIAMRoleConfig{
				Name: "{{.nuon.install.name}}-provision",
			},
			MaintenanceRole: app.AppAWSIAMRoleConfig{
				Name: "{{.nuon.install.name}}-maintenance",
			},
			DeprovisionRole: app.AppAWSIAMRoleConfig{
				Name: "{{.nuon.install.name}}-deprovision",
			},
			CustomRoles: []app.AppAWSIAMRoleConfig{
				{Name: "{{.nuon.install.id}}-custom-db-role"},
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
			ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/test-install-provision",
			MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/test-install-maintenance",
			DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/test-install-deprovision",
			CustomRoleARNs: map[string]string{
				"1234-custom-db-role": "arn:aws:iam::123456789012:role/custom-db",
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
				MatrixRules: []*app.AppOperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "maintenance"},
				},
				DefaultRole:  "provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
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
					app.OperationDeploy: "{{.nuon.install.id}}-custom-db-role",
				},
				MatrixRules: []*app.AppOperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "{{.nuon.install.name}}-maintenance"},
				},
				DefaultRole:  "{{.nuon.install.name}}-provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
			},
			expectedRoleName: "1234-custom-db-role",
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
				MatrixRules: []*app.AppOperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "{{.nuon.install.name}}-maintenance"},
				},
				DefaultRole:  "{{.nuon.install.name}}-provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
			},
			expectedRoleName: "test-install-maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/test-install-maintenance",
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
				DefaultRole:   "{{.nuon.install.name}}-provision",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
				InstallState:  baseInstallState,
			},
			expectedRoleName: "test-install-provision",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/test-install-provision",
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
				DefaultRole:   "{{.nuon.install.name}}-provision",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
				InstallState:  baseInstallState,
			},
			expectedRoleName: "test-install-provision",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/test-install-provision",
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
				DefaultRole:   "{{.nuon.install.name}}-deprovision",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
				InstallState:  baseInstallState,
			},
			expectedRoleName: "test-install-deprovision",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/test-install-deprovision",
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
				DefaultRole:   "{{.nuon.install.name}}-maintenance",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
				InstallState:  baseInstallState,
			},
			expectedRoleName: "test-install-maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/test-install-maintenance",
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
				InstallState:  baseInstallState,
			},
			expectError:   true,
			errorContains: "role nonexistent-role not found",
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
				InstallState:  baseInstallState,
			},
			expectError:   true,
			errorContains: "role invalid-role not found",
		},
		{
			name: "wildcard matrix rule matches component",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "any-component",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules: []*app.AppOperationRoleRule{
					{PrincipalType: "component", PrincipalName: "*", Operation: app.OperationDeploy, Role: "{{.nuon.install.name}}-maintenance"},
				},
				DefaultRole:  "{{.nuon.install.name}}-provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
			},
			expectedRoleName: "test-install-maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/test-install-maintenance",
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
				MatrixRules: []*app.AppOperationRoleRule{
					{PrincipalType: "action", PrincipalName: "*", Operation: app.OperationTrigger, Role: "emergency-access"},
				},
				DefaultRole:  "maintenance",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
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
				InstallState: baseInstallState,
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
				MatrixRules: []*app.AppOperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "maintenance"},
				},
				DefaultRole: "maintenance",
				AppConfig:   baseAppConfig,
				StackOutputs: &app.InstallStackOutputs{
					AzureStackOutputs: &app.AzureStackOutputs{
						SubscriptionID: "test-subscription-id",
					},
				},
				InstallState: baseInstallState,
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
				DefaultRole:   "{{.nuon.install.name}}-provision",
				AppConfig:     baseAppConfig,
				StackOutputs:  baseStackOutputs,
				InstallState:  baseInstallState,
			},
			expectedRoleName: "test-install-provision",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/test-install-provision",
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
				InstallState:  baseInstallState,
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
					app.OperationDeploy: "{{.nuon.install.id}}-custom-db-role",
				},
				MatrixRules:  nil,
				DefaultRole:  "{{.nuon.install.name}}-maintenance",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
			},
			expectedRoleName: "test-install-maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/test-install-maintenance",
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
				MatrixRules: []*app.AppOperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "{{.nuon.install.id}}-custom-db-role"},
				},
				DefaultRole:  "{{.nuon.install.name}}-maintenance",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
			},
			expectedRoleName: "test-install-maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/test-install-maintenance",
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
				RuntimeRole:   "1234-custom-db-role",
				EntityRoles: EntityOperationRoleMap{
					app.OperationDeploy: "emergency-access",
				},
				MatrixRules: []*app.AppOperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "{{.nuon.install.name}}-maintenance"},
				},
				DefaultRole:  "{{.nuon.install.name}}-provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
			},
			expectedRoleName: "1234-custom-db-role",
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
				DefaultRole:    "{{.nuon.install.name}}-maintenance",
				AppConfig:      baseAppConfig,
				StackOutputs:   baseStackOutputs,
				InstallState:   baseInstallState,
			},
			expectedRoleName: "test-install-maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/test-install-maintenance",
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
				InstallState:   baseInstallState,
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
					app.OperationDeploy: "{{.nuon.install.id}}-custom-db-role",
				},
				MatrixRules:  nil,
				DefaultRole:  "{{.nuon.install.name}}-maintenance",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
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
				MatrixRules: []*app.AppOperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "{{.nuon.install.name}}-maintenance"},
				},
				DefaultRole:  "{{.nuon.install.name}}-provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
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
				DefaultRole:    "{{.nuon.install.name}}-maintenance",
				AppConfig:      baseAppConfig,
				StackOutputs:   baseStackOutputs,
				InstallState:   baseInstallState,
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
				RuntimeRole:    "1234-custom-db-role",
				BreakGlassRole: "emergency-access",
				EntityRoles:    nil,
				MatrixRules:    nil,
				DefaultRole:    "{{.nuon.install.name}}-maintenance",
				AppConfig:      baseAppConfig,
				StackOutputs:   baseStackOutputs,
				InstallState:   baseInstallState,
			},
			expectedRoleName: "1234-custom-db-role",
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
				InstallState:   baseInstallState,
			},
			expectError:   true,
			errorContains: "role nonexistent-break-glass not found",
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
					app.OperationDeploy: "{{.nuon.install.id}}-custom-db-role",
				},
				MatrixRules: []*app.AppOperationRoleRule{
					{PrincipalType: "component", PrincipalName: "database", Operation: app.OperationDeploy, Role: "{{.nuon.install.name}}-maintenance"},
				},
				DefaultRole:  "{{.nuon.install.name}}-provision",
				AppConfig:    baseAppConfig,
				StackOutputs: baseStackOutputs,
				InstallState: baseInstallState,
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
				InstallState: baseInstallState,
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
				InstallState: baseInstallState,
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

func TestSelectRoleWithTemplateRendering(t *testing.T) {
	// Create install state with realistic values for template rendering
	installState := &state.State{
		Install: &state.InstallState{
			ID:   "ins_abc123xyz",
			Name: "production",
		},
		Org: &state.OrgState{
			ID:   "org_def456uvw",
			Name: "acme-corp",
		},
	}

	// AppConfig with templated role names (realistic production scenario)
	templatedAppConfig := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			ProvisionRole: app.AppAWSIAMRoleConfig{
				Name: "{{.nuon.install.id}}-provision",
			},
			MaintenanceRole: app.AppAWSIAMRoleConfig{
				Name: "{{.nuon.install.name}}-maintenance",
			},
			DeprovisionRole: app.AppAWSIAMRoleConfig{
				Name: "{{.nuon.org.name}}-deprovision",
			},
			CustomRoles: []app.AppAWSIAMRoleConfig{
				{Name: "{{.nuon.install.id}}-custom-db"},
				{Name: "{{.nuon.install.name}}-custom-api"},
			},
		},
		BreakGlassConfig: app.AppBreakGlassConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{Name: "{{.nuon.install.id}}-emergency"},
			},
		},
	}

	// Stack outputs with rendered role ARNs matching the templates above
	templatedStackOutputs := &app.InstallStackOutputs{
		AWSStackOutputs: &app.AWSStackOutputs{
			ProvisionIAMRoleARN:   "arn:aws:iam::123456789012:role/ins_abc123xyz-provision",
			MaintenanceIAMRoleARN: "arn:aws:iam::123456789012:role/production-maintenance",
			DeprovisionIAMRoleARN: "arn:aws:iam::123456789012:role/acme-corp-deprovision",
			CustomRoleARNs: map[string]string{
				"ins_abc123xyz-custom-db": "arn:aws:iam::123456789012:role/ins_abc123xyz-custom-db",
				"production-custom-api":   "arn:aws:iam::123456789012:role/production-custom-api",
			},
			BreakGlassRoleARNs: map[string]string{
				"ins_abc123xyz-emergency": "arn:aws:iam::123456789012:role/ins_abc123xyz-emergency",
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
	}{
		{
			name: "default provision role with install.id template renders correctly",
			ctx: &SelectionContext{
				Operation:     app.OperationProvision,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "api",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "{{.nuon.install.id}}-provision",
				AppConfig:     templatedAppConfig,
				StackOutputs:  templatedStackOutputs,
				InstallState:  installState,
			},
			expectedRoleName: "ins_abc123xyz-provision",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/ins_abc123xyz-provision",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "default maintenance role with install.name template renders correctly",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "api",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "{{.nuon.install.name}}-maintenance",
				AppConfig:     templatedAppConfig,
				StackOutputs:  templatedStackOutputs,
				InstallState:  installState,
			},
			expectedRoleName: "production-maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/production-maintenance",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "default deprovision role with org.name template renders correctly",
			ctx: &SelectionContext{
				Operation:     app.OperationDeprovision,
				PrincipalType: principal.TypeSandbox,
				PrincipalName: "",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "{{.nuon.org.name}}-deprovision",
				AppConfig:     templatedAppConfig,
				StackOutputs:  templatedStackOutputs,
				InstallState:  installState,
			},
			expectedRoleName: "acme-corp-deprovision",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/acme-corp-deprovision",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "runtime role with template renders correctly",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "{{.nuon.install.id}}-emergency",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "{{.nuon.install.id}}-provision",
				AppConfig:     templatedAppConfig,
				StackOutputs:  templatedStackOutputs,
				InstallState:  installState,
			},
			expectedRoleName: "ins_abc123xyz-emergency",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/ins_abc123xyz-emergency",
			expectedSource:   RoleSelectionSourceRuntime,
			expectError:      false,
		},
		{
			name: "entity role with template renders correctly",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "database",
				RuntimeRole:   "",
				EntityRoles: EntityOperationRoleMap{
					app.OperationDeploy: "{{.nuon.install.id}}-custom-db",
				},
				MatrixRules:  nil,
				DefaultRole:  "{{.nuon.install.id}}-provision",
				AppConfig:    templatedAppConfig,
				StackOutputs: templatedStackOutputs,
				InstallState: installState,
			},
			expectedRoleName: "ins_abc123xyz-custom-db",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/ins_abc123xyz-custom-db",
			expectedSource:   RoleSelectionSourceEntity,
			expectError:      false,
		},
		{
			name: "matrix rule role with template renders correctly",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "api",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules: []*app.AppOperationRoleRule{
					{
						PrincipalType: principal.TypeComponent,
						PrincipalName: "api",
						Operation:     app.OperationDeploy,
						Role:          "{{.nuon.install.name}}-custom-api",
					},
				},
				DefaultRole:  "{{.nuon.install.id}}-provision",
				AppConfig:    templatedAppConfig,
				StackOutputs: templatedStackOutputs,
				InstallState: installState,
			},
			expectedRoleName: "production-custom-api",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/production-custom-api",
			expectedSource:   RoleSelectionSourceMatrix,
			expectError:      false,
		},
		{
			name: "break glass role with template renders correctly",
			ctx: &SelectionContext{
				Operation:      app.OperationTrigger,
				PrincipalType:  principal.TypeAction,
				PrincipalName:  "emergency-deploy",
				RuntimeRole:    "",
				BreakGlassRole: "{{.nuon.install.id}}-emergency",
				EntityRoles:    nil,
				MatrixRules:    nil,
				DefaultRole:    "{{.nuon.install.id}}-provision",
				AppConfig:      templatedAppConfig,
				StackOutputs:   templatedStackOutputs,
				InstallState:   installState,
			},
			expectedRoleName: "ins_abc123xyz-emergency",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/ins_abc123xyz-emergency",
			expectedSource:   RoleSelectionSourceBreakGlass,
			expectError:      false,
		},
		{
			name: "sandbox mode with templated default role renders correctly",
			ctx: &SelectionContext{
				SandboxMode:   true,
				Operation:     app.OperationProvision,
				PrincipalType: principal.TypeSandbox,
				PrincipalName: "",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "{{.nuon.install.name}}-maintenance",
				AppConfig:     templatedAppConfig,
				StackOutputs:  templatedStackOutputs,
				InstallState:  installState,
			},
			expectedRoleName: "production-maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/production-maintenance",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "complex template with multiple variables renders correctly",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "api",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "{{.nuon.install.name}}-maintenance",
				AppConfig:     templatedAppConfig,
				StackOutputs:  templatedStackOutputs,
				InstallState:  installState,
			},
			expectedRoleName: "production-maintenance",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/production-maintenance",
			expectedSource:   RoleSelectionSourceDefault,
			expectError:      false,
		},
		{
			name: "template rendering fails gracefully with invalid template",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "api",
				RuntimeRole:   "",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "{{.nuon.install.name",
				AppConfig:     templatedAppConfig,
				StackOutputs:  templatedStackOutputs,
				InstallState:  installState,
			},
			expectError: true,
		},
		{
			name: "mixed: template in DefaultRole, plain text in RuntimeRole",
			ctx: &SelectionContext{
				Operation:     app.OperationDeploy,
				PrincipalType: principal.TypeComponent,
				PrincipalName: "api",
				RuntimeRole:   "ins_abc123xyz-emergency",
				EntityRoles:   nil,
				MatrixRules:   nil,
				DefaultRole:   "{{.nuon.install.id}}-provision",
				AppConfig:     templatedAppConfig,
				StackOutputs:  templatedStackOutputs,
				InstallState:  installState,
			},
			expectedRoleName: "ins_abc123xyz-emergency",
			expectedRoleARN:  "arn:aws:iam::123456789012:role/ins_abc123xyz-emergency",
			expectedSource:   RoleSelectionSourceRuntime,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SelectRole(tt.ctx, zap.NewNop())

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

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
