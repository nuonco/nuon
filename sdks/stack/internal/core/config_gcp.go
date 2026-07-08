package core

// GCPConfig carries the GCP-specific inputs for an install stack. It mirrors
// the GCP-specific variables in install-stacks/gcp/variables.tf (pinned to the
// map-based break_glass_roles/custom_roles shape). Populated on Config.GCP
// when Config.Cloud is gcp.
//
// Fields split by who populates them: the Nuon-generated set comes from ctl-api
// alongside the stack run; the customer-supplied set (ProjectID, Region,
// RunnerMachineType, HasGKENodePool, GKENodePoolSAEmail) is filled by the SDK
// from Options at provision time, mirroring how AWS sources its region.
type GCPConfig struct {
	// Customer-supplied at provision time. ProjectID and Region are required
	// by the module (no default); the GKE and machine-type fields fall back to
	// the module's defaults when left unset.
	ProjectID          string `json:"project_id,omitempty"`
	Region             string `json:"region,omitempty"`
	RunnerMachineType  string `json:"runner_machine_type,omitempty"`
	HasGKENodePool     *bool  `json:"has_gke_node_pool,omitempty"`
	GKENodePoolSAEmail string `json:"gke_node_pool_sa_email,omitempty"`

	// Nuon-generated: runner bootstrap. RunnerInitScriptURL is required by the
	// module; RunnerAPIToken is optional (init-mng-v2 fetches its own token).
	RunnerInitScriptURL string `json:"runner_init_script_url,omitempty"`
	RunnerAPIToken      string `json:"runner_api_token,omitempty"`

	// Nuon-generated: operation roles. Unlike AWS (inline policy docs + managed
	// ARNs), GCP binds a list of IAM permissions (custom role) and/or a single
	// predefined role per service account.
	ProvisionPermissions      []string `json:"provision_permissions,omitempty"`
	ProvisionPredefinedRole   string   `json:"provision_predefined_role,omitempty"`
	MaintenancePermissions    []string `json:"maintenance_permissions,omitempty"`
	MaintenancePredefinedRole string   `json:"maintenance_predefined_role,omitempty"`
	DeprovisionPermissions    []string `json:"deprovision_permissions,omitempty"`
	DeprovisionPredefinedRole string   `json:"deprovision_predefined_role,omitempty"`

	// BreakGlassRoles / CustomRoles are keyed by the role name to use verbatim,
	// matching the module's map(object) variables.
	BreakGlassRoles map[string]GCPRole `json:"break_glass_roles,omitempty"`
	CustomRoles     map[string]GCPRole `json:"custom_roles,omitempty"`
}

// GCPRole is the per-role payload for GCP break-glass/custom roles. It mirrors
// the module's object type: a custom-role permission list and/or a predefined
// role, plus an enabled flag.
type GCPRole struct {
	Permissions    []string `json:"permissions,omitempty"`
	PredefinedRole string   `json:"predefined_role,omitempty"`
	Enabled        bool     `json:"enabled,omitempty"`
}
