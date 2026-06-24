package terraform

import (
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// gcpTFRole mirrors the module's break_glass_roles / custom_roles object type.
// Every attribute is emitted unconditionally (no omitempty) because Terraform's
// object() type rejects a value missing any attribute.
type gcpTFRole struct {
	Permissions    []string `json:"permissions"`
	PredefinedRole string   `json:"predefined_role"`
	Enabled        bool     `json:"enabled"`
}

// gcpTFVars mirrors install-stacks/gcp/variables.tf. Field json tags are the
// exact Terraform variable names. Nuon-generated variables are emitted
// unconditionally; customer-supplied ones (project/region/machine-type/GKE)
// are omitempty so unset values fall back to the module's defaults — except
// gcp_project_id / gcp_region, which the module requires and the SDK must
// supply before apply.
//
// Secrets reuse tfSecret (same {description, required, value} shape as AWS).
type gcpTFVars struct {
	NuonInstallID string `json:"nuon_install_id"`
	NuonOrgID     string `json:"nuon_org_id"`
	NuonAppID     string `json:"nuon_app_id"`

	RunnerAPIURL        string `json:"runner_api_url"`
	RunnerAPIToken      string `json:"runner_api_token,omitempty"`
	RunnerID            string `json:"runner_id"`
	RunnerInitScriptURL string `json:"runner_init_script_url"`
	PhoneHomeURL        string `json:"phone_home_url"`

	ProvisionPermissions      []string `json:"provision_permissions"`
	ProvisionPredefinedRole   string   `json:"provision_predefined_role"`
	MaintenancePermissions    []string `json:"maintenance_permissions"`
	MaintenancePredefinedRole string   `json:"maintenance_predefined_role"`
	DeprovisionPermissions    []string `json:"deprovision_permissions"`
	DeprovisionPredefinedRole string   `json:"deprovision_predefined_role"`

	BreakGlassRoles map[string]gcpTFRole `json:"break_glass_roles"`
	CustomRoles     map[string]gcpTFRole `json:"custom_roles"`

	InstallInputs       map[string]string   `json:"install_inputs"`
	AutoGenerateSecrets []string            `json:"auto_generate_secrets"`
	Secrets             map[string]tfSecret `json:"secrets"`

	GCPProjectID       string `json:"gcp_project_id,omitempty"`
	GCPRegion          string `json:"gcp_region,omitempty"`
	RunnerMachineType  string `json:"runner_machine_type,omitempty"`
	HasGKENodePool     *bool  `json:"has_gke_node_pool,omitempty"`
	GKENodePoolSAEmail string `json:"gke_node_pool_sa_email,omitempty"`
}

func renderGCPTFVars(cfg *core.Config) ([]byte, error) {
	if cfg.GCP == nil {
		return nil, fmt.Errorf("gcp terraform: config missing gcp block")
	}
	g := cfg.GCP
	v := gcpTFVars{
		NuonInstallID: cfg.InstallID,
		NuonOrgID:     cfg.OrgID,
		NuonAppID:     cfg.AppID,

		RunnerAPIURL:        cfg.RunnerAPIURL,
		RunnerAPIToken:      g.RunnerAPIToken,
		RunnerID:            cfg.RunnerID,
		RunnerInitScriptURL: g.RunnerInitScriptURL,
		PhoneHomeURL:        cfg.PhoneHomeURL,

		ProvisionPermissions:      nonNilSlice(g.ProvisionPermissions),
		ProvisionPredefinedRole:   g.ProvisionPredefinedRole,
		MaintenancePermissions:    nonNilSlice(g.MaintenancePermissions),
		MaintenancePredefinedRole: g.MaintenancePredefinedRole,
		DeprovisionPermissions:    nonNilSlice(g.DeprovisionPermissions),
		DeprovisionPredefinedRole: g.DeprovisionPredefinedRole,

		BreakGlassRoles: nonNilGCPRoles(g.BreakGlassRoles),
		CustomRoles:     nonNilGCPRoles(g.CustomRoles),

		InstallInputs:       nonNilStrMap(cfg.InstallInputs),
		AutoGenerateSecrets: nonNilSlice(cfg.AutoGenerateSecrets),
		Secrets:             nonNilSecrets(cfg.Secrets),

		GCPProjectID:       g.ProjectID,
		GCPRegion:          g.Region,
		RunnerMachineType:  g.RunnerMachineType,
		HasGKENodePool:     g.HasGKENodePool,
		GKENodePoolSAEmail: g.GKENodePoolSAEmail,
	}
	return json.MarshalIndent(v, "", "  ")
}

func nonNilGCPRoles(m map[string]core.GCPRole) map[string]gcpTFRole {
	out := make(map[string]gcpTFRole, len(m))
	for k, v := range m {
		out[k] = gcpTFRole{
			Permissions:    nonNilSlice(v.Permissions),
			PredefinedRole: v.PredefinedRole,
			Enabled:        v.Enabled,
		}
	}
	return out
}
