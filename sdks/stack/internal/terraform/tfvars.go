package terraform

import (
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// tfRole / tfSecret mirror the module's object() variable attributes. We do
// NOT reuse core.RoleConfig / core.SecretInput directly: their json tags carry
// `omitempty`, which would drop attributes the module's object type requires
// (role `enabled`, and all three secret attributes), causing Terraform to
// reject the value. These structs emit every attribute unconditionally.
type tfRole struct {
	Permissions          []string `json:"permissions"`
	InlinePolicyDocument string   `json:"inline_policy_document"`
	ManagedPolicyARNs    []string `json:"managed_policy_arns"`
	Enabled              bool     `json:"enabled"`
}

type tfSecret struct {
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Value       string `json:"value"`
}

// tfvars mirrors install-stacks/aws/variables.tf. Field json tags are the
// exact Terraform variable names.
//
// We emit `.tfvars.json` rather than HCL so inline IAM policy documents and
// the like need no escaping.
type tfvars struct {
	NuonInstallID string `json:"nuon_install_id"`
	NuonOrgID     string `json:"nuon_org_id"`
	NuonAppID     string `json:"nuon_app_id"`
	AWSRegion     string `json:"aws_region"`
	RunnerAPIURL  string `json:"runner_api_url"`
	RunnerID      string `json:"runner_id"`

	// PhoneHomeURL drives the module's phone-home: the module reports the run
	// to this endpoint, so the SDK does not report it itself (see
	// Provisioner.ReportsOwnRun).
	PhoneHomeURL string `json:"phone_home_url"`

	NuonSupportIAMRoleARNs []string `json:"nuon_support_iam_role_arns"`

	ProvisionPermissions          []string `json:"provision_permissions"`
	ProvisionInlinePolicyDocument string   `json:"provision_inline_policy_document"`
	ProvisionManagedPolicyARNs    []string `json:"provision_managed_policy_arns"`

	MaintenancePermissions          []string `json:"maintenance_permissions"`
	MaintenanceInlinePolicyDocument string   `json:"maintenance_inline_policy_document"`
	MaintenanceManagedPolicyARNs    []string `json:"maintenance_managed_policy_arns"`

	DeprovisionPermissions          []string `json:"deprovision_permissions"`
	DeprovisionInlinePolicyDocument string   `json:"deprovision_inline_policy_document"`
	DeprovisionManagedPolicyARNs    []string `json:"deprovision_managed_policy_arns"`

	BreakGlassRoles map[string]tfRole `json:"break_glass_roles"`
	CustomRoles     map[string]tfRole `json:"custom_roles"`

	InstallInputs map[string]string `json:"install_inputs"`

	AutoGenerateSecrets []string            `json:"auto_generate_secrets"`
	Secrets             map[string]tfSecret `json:"secrets"`
}

// renderTFVars builds the terraform.tfvars.json bytes from cfg. Slices/maps
// are normalized to non-nil so the JSON emits [] / {} (Terraform treats those
// as set; a null would fall back to the variable default, which is fine here
// but we prefer to be explicit).
func renderTFVars(cfg *core.Config) ([]byte, error) {
	if cfg.AWS == nil {
		return nil, fmt.Errorf("aws terraform: config missing aws block")
	}
	awsCfg := cfg.AWS
	v := tfvars{
		NuonInstallID: cfg.InstallID,
		NuonOrgID:     cfg.OrgID,
		NuonAppID:     cfg.AppID,
		AWSRegion:     awsCfg.Region,
		RunnerAPIURL:  cfg.RunnerAPIURL,
		RunnerID:      cfg.RunnerID,
		PhoneHomeURL:  cfg.PhoneHomeURL,

		NuonSupportIAMRoleARNs: nonNilSlice(awsCfg.NuonSupportIAMRoleARNs),

		ProvisionPermissions:          nonNilSlice(awsCfg.ProvisionPermissions),
		ProvisionInlinePolicyDocument: awsCfg.ProvisionInlinePolicyDocument,
		ProvisionManagedPolicyARNs:    nonNilSlice(awsCfg.ProvisionManagedPolicyARNs),

		MaintenancePermissions:          nonNilSlice(awsCfg.MaintenancePermissions),
		MaintenanceInlinePolicyDocument: awsCfg.MaintenanceInlinePolicyDocument,
		MaintenanceManagedPolicyARNs:    nonNilSlice(awsCfg.MaintenanceManagedPolicyARNs),

		DeprovisionPermissions:          nonNilSlice(awsCfg.DeprovisionPermissions),
		DeprovisionInlinePolicyDocument: awsCfg.DeprovisionInlinePolicyDocument,
		DeprovisionManagedPolicyARNs:    nonNilSlice(awsCfg.DeprovisionManagedPolicyARNs),

		BreakGlassRoles: nonNilRoles(awsCfg.BreakGlassRoles),
		CustomRoles:     nonNilRoles(awsCfg.CustomRoles),

		InstallInputs: nonNilStrMap(cfg.InstallInputs),

		AutoGenerateSecrets: nonNilSlice(cfg.AutoGenerateSecrets),
		Secrets:             nonNilSecrets(cfg.Secrets),
	}
	return json.MarshalIndent(v, "", "  ")
}

func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func nonNilStrMap(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}

func nonNilRoles(m map[string]core.RoleConfig) map[string]tfRole {
	out := make(map[string]tfRole, len(m))
	for k, v := range m {
		out[k] = tfRole{
			Permissions:          nonNilSlice(v.Permissions),
			InlinePolicyDocument: v.InlinePolicyDocument,
			ManagedPolicyARNs:    nonNilSlice(v.ManagedPolicyARNs),
			Enabled:              v.Enabled,
		}
	}
	return out
}

func nonNilSecrets(m map[string]core.SecretInput) map[string]tfSecret {
	out := make(map[string]tfSecret, len(m))
	for k, v := range m {
		out[k] = tfSecret{
			Description: v.Description,
			Required:    v.Required,
			Value:       v.Value,
		}
	}
	return out
}
