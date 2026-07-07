package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	stack "github.com/nuonco/nuon/sdks/stack"
)

// stackDataSourceModel is the Terraform shape for the nuon_stack data source. It
// mirrors the install-stack config the control plane renders (the same data the
// legacy flow wrote to tfvars), so an install-stacks module can read it from the
// API instead of receiving it as variables.
type stackDataSourceModel struct {
	PhoneHomeID types.String `tfsdk:"phone_home_id"`

	InstallID    types.String `tfsdk:"install_id"`
	OrgID        types.String `tfsdk:"org_id"`
	AppID        types.String `tfsdk:"app_id"`
	Cloud        types.String `tfsdk:"cloud"`
	RunnerID     types.String `tfsdk:"runner_id"`
	RunnerAPIURL types.String `tfsdk:"runner_api_url"`
	PhoneHomeURL types.String `tfsdk:"phone_home_url"`

	InstallInputs       map[string]string   `tfsdk:"install_inputs"`
	AutoGenerateSecrets []string            `tfsdk:"auto_generate_secrets"`
	Secrets             map[string]secretTF `tfsdk:"secrets"`

	GCP *gcpTF `tfsdk:"gcp"`
}

// secretTF mirrors the module's secrets map(object) element.
type secretTF struct {
	Description string `tfsdk:"description"`
	Required    bool   `tfsdk:"required"`
	Value       string `tfsdk:"value"`
}

// gcpRoleTF mirrors the module's break_glass_roles / custom_roles element.
type gcpRoleTF struct {
	Permissions    []string `tfsdk:"permissions"`
	PredefinedRole string   `tfsdk:"predefined_role"`
	Enabled        bool     `tfsdk:"enabled"`
}

// gcpTF carries the GCP-specific install-stack config.
type gcpTF struct {
	RunnerInitScriptURL string `tfsdk:"runner_init_script_url"`
	RunnerAPIToken      string `tfsdk:"runner_api_token"`

	ProvisionPermissions      []string `tfsdk:"provision_permissions"`
	ProvisionPredefinedRole   string   `tfsdk:"provision_predefined_role"`
	MaintenancePermissions    []string `tfsdk:"maintenance_permissions"`
	MaintenancePredefinedRole string   `tfsdk:"maintenance_predefined_role"`
	DeprovisionPermissions    []string `tfsdk:"deprovision_permissions"`
	DeprovisionPredefinedRole string   `tfsdk:"deprovision_predefined_role"`

	BreakGlassRoles map[string]gcpRoleTF `tfsdk:"break_glass_roles"`
	CustomRoles     map[string]gcpRoleTF `tfsdk:"custom_roles"`
}

// flattenConfig copies the fetched SDK config onto the data source model,
// preserving the caller-supplied phone_home_id.
func flattenConfig(data *stackDataSourceModel, cfg *stack.Config) {
	data.InstallID = types.StringValue(cfg.InstallID)
	data.OrgID = types.StringValue(cfg.OrgID)
	data.AppID = types.StringValue(cfg.AppID)
	data.Cloud = types.StringValue(string(cfg.Cloud))
	data.RunnerID = types.StringValue(cfg.RunnerID)
	data.RunnerAPIURL = types.StringValue(cfg.RunnerAPIURL)
	data.PhoneHomeURL = types.StringValue(cfg.PhoneHomeURL)

	// Collections are emitted as empty (never null) so module authors can call
	// length()/for_each on them without coalescing — matching the contract the
	// legacy tfvars flow provided via variable defaults.
	data.InstallInputs = orEmptyMap(cfg.InstallInputs)
	data.AutoGenerateSecrets = orEmptySlice(cfg.AutoGenerateSecrets)

	data.Secrets = make(map[string]secretTF, len(cfg.Secrets))
	for name, s := range cfg.Secrets {
		data.Secrets[name] = secretTF{
			Description: s.Description,
			Required:    s.Required,
			Value:       s.Value,
		}
	}

	if cfg.GCP != nil {
		data.GCP = flattenGCP(cfg.GCP)
	}
}

func flattenGCP(g *stack.GCPConfig) *gcpTF {
	return &gcpTF{
		RunnerInitScriptURL:       g.RunnerInitScriptURL,
		RunnerAPIToken:            g.RunnerAPIToken,
		ProvisionPermissions:      orEmptySlice(g.ProvisionPermissions),
		ProvisionPredefinedRole:   g.ProvisionPredefinedRole,
		MaintenancePermissions:    orEmptySlice(g.MaintenancePermissions),
		MaintenancePredefinedRole: g.MaintenancePredefinedRole,
		DeprovisionPermissions:    orEmptySlice(g.DeprovisionPermissions),
		DeprovisionPredefinedRole: g.DeprovisionPredefinedRole,
		BreakGlassRoles:           flattenGCPRoles(g.BreakGlassRoles),
		CustomRoles:               flattenGCPRoles(g.CustomRoles),
	}
}

func flattenGCPRoles(in map[string]stack.GCPRole) map[string]gcpRoleTF {
	out := make(map[string]gcpRoleTF, len(in))
	for name, r := range in {
		out[name] = gcpRoleTF{
			Permissions:    orEmptySlice(r.Permissions),
			PredefinedRole: r.PredefinedRole,
			Enabled:        r.Enabled,
		}
	}
	return out
}

func orEmptySlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func orEmptyMap(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}
