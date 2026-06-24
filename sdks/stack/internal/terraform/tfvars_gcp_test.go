package terraform

import (
	"encoding/json"
	"testing"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// TestRenderGCPTFVarsKeys locks the tfvars variable names to
// install-stacks/gcp variables.tf. Drift here silently mis-feeds the module.
func TestRenderGCPTFVarsKeys(t *testing.T) {
	b, err := renderGCPTFVars(&core.Config{
		Cloud:        core.CloudGCP,
		InstallID:    "inst123",
		OrgID:        "org123",
		AppID:        "app123",
		RunnerID:     "rnr123",
		RunnerAPIURL: "https://api.nuon.co/runner",
		PhoneHomeURL: "https://api.nuon.co/v1/installs/inst123/phone-home/ph",
		GCP: &core.GCPConfig{
			ProjectID:               "my-proj",
			Region:                  "us-central1",
			RunnerMachineType:       "e2-medium",
			RunnerInitScriptURL:     "https://example/init.sh",
			RunnerAPIToken:          "tok123",
			ProvisionPredefinedRole: "roles/owner",
			CustomRoles: map[string]core.GCPRole{
				"inst123-certs": {Permissions: []string{"dns.changes.create"}, PredefinedRole: "roles/certificatemanager.editor", Enabled: true},
			},
		},
	})
	if err != nil {
		t.Fatalf("renderGCPTFVars: %v", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal tfvars: %v", err)
	}

	want := []string{
		"nuon_install_id", "nuon_org_id", "nuon_app_id",
		"runner_api_url", "runner_api_token", "runner_id", "runner_init_script_url", "phone_home_url",
		"provision_permissions", "provision_predefined_role",
		"maintenance_permissions", "maintenance_predefined_role",
		"deprovision_permissions", "deprovision_predefined_role",
		"break_glass_roles", "custom_roles", "install_inputs",
		"auto_generate_secrets", "secrets",
		"gcp_project_id", "gcp_region", "runner_machine_type",
	}
	for _, k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("tfvars missing key %q", k)
		}
	}

	// has_gke_node_pool must be omitted when unset so the module default
	// (true) applies, rather than forcing a value.
	if _, ok := got["has_gke_node_pool"]; ok {
		t.Errorf("has_gke_node_pool should be omitted when nil")
	}

	// custom_roles object attribute names must match the module's object type.
	var roles map[string]map[string]json.RawMessage
	if err := json.Unmarshal(got["custom_roles"], &roles); err != nil {
		t.Fatalf("custom_roles: %v", err)
	}
	for _, attr := range []string{"permissions", "predefined_role", "enabled"} {
		if _, ok := roles["inst123-certs"][attr]; !ok {
			t.Errorf("custom role missing attr %q", attr)
		}
	}
}

// TestRenderGCPTFVarsMissingBlock guards the nil-GCP path.
func TestRenderGCPTFVarsMissingBlock(t *testing.T) {
	if _, err := renderGCPTFVars(&core.Config{Cloud: core.CloudGCP}); err == nil {
		t.Fatal("expected error when GCP config block is missing")
	}
}
