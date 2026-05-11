package stack

// Config carries the per-install rendered configuration that the ctl-api
// produces alongside a stack run. Mirrors the tfvars contract at
// install-stacks/aws/variables.tf so the SDK and Terraform paths produce
// the same set of AWS resources.
type Config struct {
	// InstallID is duplicated here so config can be threaded through the
	// stack package without dragging State along. State.InstallID remains
	// the source of truth at runtime; this field is convenience only.
	InstallID string `json:"install_id,omitempty"`
	OrgID     string `json:"org_id,omitempty"`
	AppID     string `json:"app_id,omitempty"`
	AWSRegion string `json:"aws_region,omitempty"`

	// ClusterName resolves the EKS cluster-name tag value. Mirrors
	// services/ctl-api/internal/pkg/stacks/cloudformation/nested_template_vpc.go
	// getClusterName: install input "cluster_name" if set, else install_id.
	ClusterName string `json:"cluster_name,omitempty"`

	RunnerID     string `json:"runner_id,omitempty"`
	RunnerAPIURL string `json:"runner_api_url,omitempty"`

	// NuonSupportIAMRoleARNs lists Nuon control-plane IAM role ARNs that may
	// assume the operation roles. Empty falls back to the customer's account
	// root, matching the TF module's control_plane_assume default.
	NuonSupportIAMRoleARNs []string `json:"nuon_support_iam_role_arns,omitempty"`

	InstallInputs map[string]string `json:"install_inputs,omitempty"`

	AutoGenerateSecrets []string               `json:"auto_generate_secrets,omitempty"`
	Secrets             map[string]SecretInput `json:"secrets,omitempty"`

	// Operation role inputs. Inline document takes precedence over Permissions.
	ProvisionPermissions          []string `json:"provision_permissions,omitempty"`
	ProvisionInlinePolicyDocument string   `json:"provision_inline_policy_document,omitempty"`
	ProvisionManagedPolicyARNs    []string `json:"provision_managed_policy_arns,omitempty"`

	MaintenancePermissions          []string `json:"maintenance_permissions,omitempty"`
	MaintenanceInlinePolicyDocument string   `json:"maintenance_inline_policy_document,omitempty"`
	MaintenanceManagedPolicyARNs    []string `json:"maintenance_managed_policy_arns,omitempty"`

	DeprovisionPermissions          []string `json:"deprovision_permissions,omitempty"`
	DeprovisionInlinePolicyDocument string   `json:"deprovision_inline_policy_document,omitempty"`
	DeprovisionManagedPolicyARNs    []string `json:"deprovision_managed_policy_arns,omitempty"`

	// BreakGlassRoles / CustomRoles are keyed by the IAM role name to use
	// verbatim — TF uses each.key directly to avoid double-prefixing past
	// IAM's 64-char limit; we follow the same contract.
	BreakGlassRoles map[string]RoleConfig `json:"break_glass_roles,omitempty"`
	CustomRoles     map[string]RoleConfig `json:"custom_roles,omitempty"`
}

// SecretInput mirrors the customer-provided secret shape.
type SecretInput struct {
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Value       string `json:"value,omitempty"`
}

// RoleConfig is the per-role payload for break-glass/custom roles.
type RoleConfig struct {
	Permissions          []string `json:"permissions,omitempty"`
	InlinePolicyDocument string   `json:"inline_policy_document,omitempty"`
	ManagedPolicyARNs    []string `json:"managed_policy_arns,omitempty"`
	Enabled              bool     `json:"enabled,omitempty"`
}

// Prefix is the resource-name prefix used across the stack. Matches the TF
// module's `local.prefix = var.nuon_install_id` exactly. Why we don't add
// "nuon-": ctl-api and downstream app templates derive role / log-group /
// secret names from the install id directly (e.g. the runner's IID
// validation expects `{install_id}-runner` as the role name); double-
// prefixing breaks every cross-system lookup.
func (c *Config) Prefix() string {
	return c.InstallID
}
