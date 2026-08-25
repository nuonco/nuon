package app

// InstallerSDKConfig is the JSON shape the stack SDK expects for a stack
// version's rendered configuration. Field tags match sdks/stack/internal/core
// (Config + per-cloud sub-structs) exactly: common fields here, cloud-specific
// inputs on AWS/GCP, selected by Cloud. Served read-only by the config endpoint
// the Terraform provider's nuon_stack data source consumes.
type InstallerSDKConfig struct {
	Cloud        string `json:"cloud,omitempty"`
	InstallID    string `json:"install_id,omitempty"`
	OrgID        string `json:"org_id,omitempty"`
	AppID        string `json:"app_id,omitempty"`
	RunnerID     string `json:"runner_id,omitempty"`
	RunnerAPIURL string `json:"runner_api_url,omitempty"`
	PhoneHomeURL string `json:"phone_home_url,omitempty"`

	InstallInputs map[string]string `json:"install_inputs,omitempty"`
	// RequiredInputs lists the install-input names that must be set before
	// provisioning. The SDK enforces it at provision time.
	RequiredInputs []string `json:"required_inputs,omitempty"`
	// SensitiveInputs lists the install-input names the app declares sensitive.
	// InstallInputs is a released map[string]string with no room for per-key
	// metadata, so sensitivity rides alongside as a name list — same shape as
	// RequiredInputs. The Terraform provider marks these values sensitive.
	SensitiveInputs []string `json:"sensitive_inputs,omitempty"`

	AutoGenerateSecrets []string                      `json:"auto_generate_secrets,omitempty"`
	Secrets             map[string]InstallerSDKSecret `json:"secrets,omitempty"`

	AWS *InstallerSDKAWSConfig `json:"aws,omitempty"`
	GCP *InstallerSDKGCPConfig `json:"gcp,omitempty"`
}

// InstallerSDKAWSConfig mirrors sdks/stack core.AWSConfig.
type InstallerSDKAWSConfig struct {
	Region            string `json:"region,omitempty"`
	ClusterName       string `json:"cluster_name,omitempty"`
	RunnerMachineType string `json:"runner_machine_type,omitempty"`

	NuonSupportIAMRoleARNs []string `json:"nuon_support_iam_role_arns,omitempty"`

	BreakGlassRoles map[string]InstallerSDKRoleConfig `json:"break_glass_roles,omitempty"`
	CustomRoles     map[string]InstallerSDKRoleConfig `json:"custom_roles,omitempty"`

	ProvisionPermissions          []string `json:"provision_permissions,omitempty"`
	ProvisionInlinePolicyDocument string   `json:"provision_inline_policy_document,omitempty"`
	ProvisionManagedPolicyARNs    []string `json:"provision_managed_policy_arns,omitempty"`

	MaintenancePermissions          []string `json:"maintenance_permissions,omitempty"`
	MaintenanceInlinePolicyDocument string   `json:"maintenance_inline_policy_document,omitempty"`
	MaintenanceManagedPolicyARNs    []string `json:"maintenance_managed_policy_arns,omitempty"`

	DeprovisionPermissions          []string `json:"deprovision_permissions,omitempty"`
	DeprovisionInlinePolicyDocument string   `json:"deprovision_inline_policy_document,omitempty"`
	DeprovisionManagedPolicyARNs    []string `json:"deprovision_managed_policy_arns,omitempty"`
}

// InstallerSDKGCPConfig mirrors sdks/stack core.GCPConfig. ctl-api populates
// the Nuon-generated fields; the customer-supplied project/region/machine-type/
// GKE inputs are filled by the SDK from CLI options, so they are left empty here.
type InstallerSDKGCPConfig struct {
	// NOTE: project + region are NOT set here — the customer supplies them at
	// provision time via the CLI. ctl-api only emits the Nuon-generated inputs.
	RunnerInitScriptURL string `json:"runner_init_script_url,omitempty"`
	RunnerAPIToken      string `json:"runner_api_token,omitempty"`
	RunnerMachineType   string `json:"runner_machine_type,omitempty"`

	ProvisionPermissions      []string `json:"provision_permissions,omitempty"`
	ProvisionPredefinedRole   string   `json:"provision_predefined_role,omitempty"`
	MaintenancePermissions    []string `json:"maintenance_permissions,omitempty"`
	MaintenancePredefinedRole string   `json:"maintenance_predefined_role,omitempty"`
	DeprovisionPermissions    []string `json:"deprovision_permissions,omitempty"`
	DeprovisionPredefinedRole string   `json:"deprovision_predefined_role,omitempty"`

	// Per-policy custom roles (policy name → permissions): one custom role per
	// policy.
	ProvisionPolicies   map[string][]string `json:"provision_policies,omitempty"`
	MaintenancePolicies map[string][]string `json:"maintenance_policies,omitempty"`
	DeprovisionPolicies map[string][]string `json:"deprovision_policies,omitempty"`

	BreakGlassRoles map[string]InstallerSDKGCPRole `json:"break_glass_roles,omitempty"`
	CustomRoles     map[string]InstallerSDKGCPRole `json:"custom_roles,omitempty"`
}

// InstallerSDKSecret is the customer-provided secret shape (cloud-agnostic).
type InstallerSDKSecret struct {
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Value       string `json:"value,omitempty"`
}

// InstallerSDKRoleConfig is the per-role payload for AWS break-glass/custom roles.
type InstallerSDKRoleConfig struct {
	Permissions          []string `json:"permissions,omitempty"`
	InlinePolicyDocument string   `json:"inline_policy_document,omitempty"`
	ManagedPolicyARNs    []string `json:"managed_policy_arns,omitempty"`
	Enabled              bool     `json:"enabled,omitempty"`
}

// InstallerSDKGCPRole is the per-role payload for GCP break-glass/custom roles.
type InstallerSDKGCPRole struct {
	Permissions    []string `json:"permissions,omitempty"`
	PredefinedRole string   `json:"predefined_role,omitempty"`
	Enabled        bool     `json:"enabled,omitempty"`

	// Per-policy custom roles (policy name → permissions).
	Policies map[string][]string `json:"policies,omitempty"`
}
