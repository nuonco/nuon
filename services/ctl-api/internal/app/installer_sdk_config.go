package app

// InstallerSDKConfig is the JSON shape the stack SDK expects for a stack
// version's rendered configuration. Field tags match sdks/stack/internal/core
// (Config + per-cloud sub-structs) exactly: common fields here, cloud-specific
// inputs on AWS/GCP/Azure, selected by Cloud. Served read-only by the config endpoint
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
	// Names the app declares sensitive. InstallInputs is a released
	// map[string]string with no room for per-key metadata, so this rides alongside
	// like RequiredInputs.
	SensitiveInputs []string `json:"sensitive_inputs,omitempty"`

	AutoGenerateSecrets []string                      `json:"auto_generate_secrets,omitempty"`
	Secrets             map[string]InstallerSDKSecret `json:"secrets,omitempty"`

	AWS   *InstallerSDKAWSConfig   `json:"aws,omitempty"`
	GCP   *InstallerSDKGCPConfig   `json:"gcp,omitempty"`
	Azure *InstallerSDKAzureConfig `json:"azure,omitempty"`

	CustomStacks            []InstallerSDKCustomStack `json:"custom_stacks,omitempty"`
	CustomStacksTemplateURL string                    `json:"custom_stacks_template_url,omitempty"`
}

// InstallerSDKCustomStack is a custom nested stack.
type InstallerSDKCustomStack struct {
	Name            string            `json:"name,omitempty"`
	Index           int               `json:"index"`
	Parameters      map[string]string `json:"parameters,omitempty"`
	Module          string            `json:"module,omitempty"`
	Outputs         map[string]string `json:"outputs,omitempty"`
	InputParameters map[string]string `json:"input_parameters,omitempty"`
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
	// Empty until the install has a recorded GCP target: unlike AWS, a GCP install
	// can be created without one, and the first provision's phone home records it.
	// The module supplies its own values for that first apply.
	ProjectID string `json:"project_id,omitempty"`
	Region    string `json:"region,omitempty"`

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

// InstallerSDKAzureConfig mirrors sdks/stack core.AzureConfig.
//
// Azure grants come on two axes and the Terraform module treats them
// differently: Actions become a subscription-scoped custom role definition (the
// module always adds */register/action to it), and BuiltInRoles become direct
// assignments at resource-group scope. Built-in roles are served as GUIDs,
// already resolved by pkg/azureroles.
//
// There is no runner API token here, unlike GCP: the Azure runner authenticates
// as its own managed identity, so the config carries the container image it
// should run instead.
type InstallerSDKAzureConfig struct {
	Location             string `json:"location,omitempty"`
	SubscriptionID       string `json:"subscription_id,omitempty"`
	SubscriptionTenantID string `json:"subscription_tenant_id,omitempty"`

	RunnerVMSize      string `json:"runner_vm_size,omitempty"`
	ContainerImageURL string `json:"container_image_url,omitempty"`
	ContainerImageTag string `json:"container_image_tag,omitempty"`

	ProvisionActions        []string `json:"provision_actions,omitempty"`
	ProvisionBuiltInRoles   []string `json:"provision_built_in_roles,omitempty"`
	MaintenanceActions      []string `json:"maintenance_actions,omitempty"`
	MaintenanceBuiltInRoles []string `json:"maintenance_built_in_roles,omitempty"`
	DeprovisionActions      []string `json:"deprovision_actions,omitempty"`
	DeprovisionBuiltInRoles []string `json:"deprovision_built_in_roles,omitempty"`

	BreakGlassRoles map[string]InstallerSDKAzureRole `json:"break_glass_roles,omitempty"`
	CustomRoles     map[string]InstallerSDKAzureRole `json:"custom_roles,omitempty"`
}

// InstallerSDKAzureRole is the per-role payload for Azure break-glass/custom
// roles.
type InstallerSDKAzureRole struct {
	Actions      []string `json:"actions,omitempty"`
	BuiltInRoles []string `json:"built_in_roles,omitempty"`
	Enabled      bool     `json:"enabled"`
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
