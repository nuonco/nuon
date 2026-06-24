package core

// AWSConfig carries the AWS-specific inputs for an install stack. It mirrors
// the AWS-specific variables in install-stacks/aws/variables.tf. Populated on
// Config.AWS when Config.Cloud is aws.
type AWSConfig struct {
	// Region is the AWS region the stack is provisioned into.
	Region string `json:"region,omitempty"`

	// ClusterName resolves the EKS cluster-name tag value. Mirrors
	// services/ctl-api/internal/pkg/stacks/cloudformation/nested_template_vpc.go
	// getClusterName: install input "cluster_name" if set, else install_id.
	ClusterName string `json:"cluster_name,omitempty"`

	// NuonSupportIAMRoleARNs lists Nuon control-plane IAM role ARNs that may
	// assume the operation roles. Empty falls back to the customer's account
	// root, matching the TF module's control_plane_assume default.
	NuonSupportIAMRoleARNs []string `json:"nuon_support_iam_role_arns,omitempty"`

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

// RoleConfig is the per-role payload for AWS break-glass/custom roles. GCP's
// equivalent (predefined-role-based) is a separate type on GCPConfig.
type RoleConfig struct {
	Permissions          []string `json:"permissions,omitempty"`
	InlinePolicyDocument string   `json:"inline_policy_document,omitempty"`
	ManagedPolicyARNs    []string `json:"managed_policy_arns,omitempty"`
	Enabled              bool     `json:"enabled,omitempty"`
}
