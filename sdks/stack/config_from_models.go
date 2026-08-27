package stack

import (
	"github.com/nuonco/nuon/sdks/stack/internal/core"
	"github.com/nuonco/nuon/sdks/stack/models"
)

// configFromModel converts the generated wire type into the SDK's working
// Config. The wire type is a strict subset: provisioning-method fields
// (Method, Terraform*, TerraformBackend) and the customer-supplied GCP inputs
// are filled from Options at provision time, so they are deliberately untouched
// here.
func configFromModel(m *models.AppInstallerSDKConfig) *core.Config {
	if m == nil {
		return nil
	}

	cfg := &core.Config{
		Cloud:               core.Cloud(m.Cloud),
		InstallID:           m.InstallID,
		OrgID:               m.OrgID,
		AppID:               m.AppID,
		RunnerID:            m.RunnerID,
		RunnerAPIURL:        m.RunnerAPIURL,
		PhoneHomeURL:        m.PhoneHomeURL,
		InstallInputs:       m.InstallInputs,
		AutoGenerateSecrets: m.AutoGenerateSecrets,
		RequiredInputs:      m.RequiredInputs,
		SensitiveInputs:     m.SensitiveInputs,
	}

	if len(m.Secrets) > 0 {
		cfg.Secrets = make(map[string]core.SecretInput, len(m.Secrets))
		for k, v := range m.Secrets {
			cfg.Secrets[k] = core.SecretInput{
				Description: v.Description,
				Required:    v.Required,
				Value:       v.Value,
			}
		}
	}

	cfg.AWS = awsConfigFromModel(m.Aws)
	cfg.GCP = gcpConfigFromModel(m.Gcp)

	return cfg
}

func awsConfigFromModel(m *models.AppInstallerSDKAWSConfig) *core.AWSConfig {
	if m == nil {
		return nil
	}

	return &core.AWSConfig{
		Region:                 m.Region,
		ClusterName:            m.ClusterName,
		RunnerMachineType:      m.RunnerMachineType,
		NuonSupportIAMRoleARNs: m.NuonSupportIamRoleArns,

		ProvisionPermissions:          m.ProvisionPermissions,
		ProvisionInlinePolicyDocument: m.ProvisionInlinePolicyDocument,
		ProvisionManagedPolicyARNs:    m.ProvisionManagedPolicyArns,

		MaintenancePermissions:          m.MaintenancePermissions,
		MaintenanceInlinePolicyDocument: m.MaintenanceInlinePolicyDocument,
		MaintenanceManagedPolicyARNs:    m.MaintenanceManagedPolicyArns,

		DeprovisionPermissions:          m.DeprovisionPermissions,
		DeprovisionInlinePolicyDocument: m.DeprovisionInlinePolicyDocument,
		DeprovisionManagedPolicyARNs:    m.DeprovisionManagedPolicyArns,

		BreakGlassRoles: roleConfigsFromModel(m.BreakGlassRoles),
		CustomRoles:     roleConfigsFromModel(m.CustomRoles),
	}
}

func roleConfigsFromModel(m map[string]models.AppInstallerSDKRoleConfig) map[string]core.RoleConfig {
	if len(m) == 0 {
		return nil
	}

	out := make(map[string]core.RoleConfig, len(m))
	for k, v := range m {
		out[k] = core.RoleConfig{
			Permissions:          v.Permissions,
			InlinePolicyDocument: v.InlinePolicyDocument,
			ManagedPolicyARNs:    v.ManagedPolicyArns,
			Enabled:              v.Enabled,
		}
	}

	return out
}

func gcpConfigFromModel(m *models.AppInstallerSDKGCPConfig) *core.GCPConfig {
	if m == nil {
		return nil
	}

	return &core.GCPConfig{
		RunnerInitScriptURL: m.RunnerInitScriptURL,
		RunnerAPIToken:      m.RunnerAPIToken,
		RunnerMachineType:   m.RunnerMachineType,

		ProvisionPermissions:      m.ProvisionPermissions,
		ProvisionPredefinedRole:   m.ProvisionPredefinedRole,
		MaintenancePermissions:    m.MaintenancePermissions,
		MaintenancePredefinedRole: m.MaintenancePredefinedRole,
		DeprovisionPermissions:    m.DeprovisionPermissions,
		DeprovisionPredefinedRole: m.DeprovisionPredefinedRole,

		ProvisionPolicies:   m.ProvisionPolicies,
		MaintenancePolicies: m.MaintenancePolicies,
		DeprovisionPolicies: m.DeprovisionPolicies,

		BreakGlassRoles: gcpRolesFromModel(m.BreakGlassRoles),
		CustomRoles:     gcpRolesFromModel(m.CustomRoles),
	}
}

func gcpRolesFromModel(m map[string]models.AppInstallerSDKGCPRole) map[string]core.GCPRole {
	if len(m) == 0 {
		return nil
	}

	out := make(map[string]core.GCPRole, len(m))
	for k, v := range m {
		out[k] = core.GCPRole{
			Permissions:    v.Permissions,
			PredefinedRole: v.PredefinedRole,
			Enabled:        v.Enabled,
			Policies:       v.Policies,
		}
	}

	return out
}
