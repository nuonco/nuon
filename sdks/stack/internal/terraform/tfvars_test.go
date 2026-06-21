package terraform

import (
	"encoding/json"
	"testing"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// TestRenderTFVarsKeys locks the tfvars variable names to install-stacks/aws
// variables.tf. A drift here silently mis-feeds the module.
func TestRenderTFVarsKeys(t *testing.T) {
	b, err := renderTFVars(&core.Config{
		InstallID:    "inst123",
		OrgID:        "org123",
		AppID:        "app123",
		AWSRegion:    "us-east-1",
		RunnerID:     "rnr123",
		RunnerAPIURL: "https://api.nuon.co",
		PhoneHomeURL: "https://api.nuon.co/v1/installs/inst123/phone-home",
		BreakGlassRoles: map[string]core.RoleConfig{
			"break-glass": {Permissions: []string{"s3:*"}, Enabled: false},
		},
		CustomRoles: map[string]core.RoleConfig{
			"custom": {InlinePolicyDocument: `{"Version":"2012-10-17"}`, Enabled: true},
		},
		Secrets: map[string]core.SecretInput{
			"db": {Description: "db pw", Required: true},
		},
	})
	if err != nil {
		t.Fatalf("renderTFVars: %v", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal tfvars: %v", err)
	}

	want := []string{
		"nuon_install_id", "nuon_org_id", "nuon_app_id", "aws_region",
		"runner_api_url", "runner_id", "phone_home_url",
		"nuon_support_iam_role_arns",
		"provision_permissions", "provision_inline_policy_document", "provision_managed_policy_arns",
		"maintenance_permissions", "maintenance_inline_policy_document", "maintenance_managed_policy_arns",
		"deprovision_permissions", "deprovision_inline_policy_document", "deprovision_managed_policy_arns",
		"break_glass_roles", "custom_roles", "install_inputs",
		"auto_generate_secrets", "secrets",
	}
	for _, k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("tfvars missing key %q", k)
		}
	}

	// phone_home_url must carry through so the module reports the run.
	var phoneHome string
	if err := json.Unmarshal(got["phone_home_url"], &phoneHome); err != nil {
		t.Fatalf("phone_home_url: %v", err)
	}
	if phoneHome != "https://api.nuon.co/v1/installs/inst123/phone-home" {
		t.Errorf("phone_home_url = %q, want the configured URL", phoneHome)
	}

	// Nested role object attribute names must match the module's object type.
	var roles map[string]map[string]json.RawMessage
	if err := json.Unmarshal(got["break_glass_roles"], &roles); err != nil {
		t.Fatalf("break_glass_roles: %v", err)
	}
	for _, attr := range []string{"permissions", "inline_policy_document", "managed_policy_arns", "enabled"} {
		if _, ok := roles["break-glass"][attr]; !ok {
			t.Errorf("break_glass role missing attr %q", attr)
		}
	}
}
