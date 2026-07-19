package arm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const (
	contributorRoleGUID = "b24988ac-6180-42a0-ab88-20f7382dd24c"
	keyVaultUserGUID    = "4633458b-17de-408a-b874-0445c86b69e6"
)

func azureRolesTemplateInput() *stacks.TemplateInput {
	inp := minimalTemplateInput()
	inp.AppCfg.PermissionsConfig.Roles = []app.AppAWSIAMRoleConfig{
		{CloudPlatform: "azure", Type: app.AWSIAMRoleTypeRunnerProvision, Name: "provision", Policies: []app.AppAWSIAMPolicyConfig{{AzureActions: []string{"Microsoft.Resources/*"}}}},
		{CloudPlatform: "azure", Type: app.AWSIAMRoleTypeRunnerMaintenance, Name: "maintenance", Policies: []app.AppAWSIAMPolicyConfig{{AzureActions: []string{"Microsoft.Compute/*"}}}},
		{CloudPlatform: "azure", Type: app.AWSIAMRoleTypeRunnerDeprovision, Name: "deprovision", Policies: []app.AppAWSIAMPolicyConfig{{AzureActions: []string{"Microsoft.Resources/*/delete"}}}},
		{CloudPlatform: "azure", Type: app.AWSIAMRoleTypeCustom, Name: "db-admin", Policies: []app.AppAWSIAMPolicyConfig{{AzureBuiltInRoles: []string{"Reader"}}}},
	}
	inp.AppCfg.BreakGlassConfig.Roles = []app.AppAWSIAMRoleConfig{
		{CloudPlatform: "azure", Type: app.AWSIAMRoleTypeBreakGlass, Name: "emergency", Policies: []app.AppAWSIAMPolicyConfig{{AzureBuiltInRoles: []string{"Owner"}}}},
	}
	return inp
}

func countResourceType(resources []any, resType string) int {
	n := 0
	for _, r := range resources {
		if m, ok := r.(map[string]any); ok && m["type"] == resType {
			n++
		}
	}
	return n
}

func TestOperationIdentities_CreatedAndStripped(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(azureRolesTemplateInput())
	if err != nil {
		t.Fatalf("getAzureTemplate returned error: %v", err)
	}

	if got := countResourceType(armTmpl.Resources, "Microsoft.ManagedIdentity/userAssignedIdentities"); got != 5 {
		t.Errorf("expected 5 user-assigned identities (3 standard + 1 custom + 1 break-glass), got %d", got)
	}

	tmplBytes, err := json.MarshalIndent(armTmpl, "", "  ")
	if err != nil {
		t.Fatalf("unable to marshal ARM template: %v", err)
	}
	body := string(tmplBytes)

	// Broad grants and the register-action deployment must be stripped off the
	// runner's system identity when operation identities exist.
	if strings.Contains(body, contributorRoleGUID) {
		t.Error("Contributor role assignment on the system identity should be stripped when operation identities are used")
	}
	if strings.Contains(body, "custom-role-deployment") {
		t.Error("legacy register-action custom role deployment should be stripped when operation identities are used")
	}

	// Key Vault Secrets User stays on the system identity for secret-sync.
	if !strings.Contains(body, keyVaultUserGUID) {
		t.Error("Key Vault Secrets User role should remain on the system identity")
	}

	// VMSS should carry both its system identity and the operation identities.
	if !strings.Contains(body, "SystemAssigned, UserAssigned") {
		t.Error("runner VMSS should have combined system + user-assigned identity when operation identities exist")
	}

	// Client IDs are surfaced via the phone-home payload.
	for _, key := range []string{
		"provision_identity_client_id",
		"maintenance_identity_client_id",
		"deprovision_identity_client_id",
		"custom_identity_client_ids",
		"break_glass_identity_client_ids",
	} {
		if !strings.Contains(body, key) {
			t.Errorf("phone-home payload missing %q", key)
		}
	}

	assertNoNestedBrackets(t, tmplBytes)
}

func TestOperationIdentities_LegacyWhenNoAzureRoles(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(minimalTemplateInput())
	if err != nil {
		t.Fatalf("getAzureTemplate returned error: %v", err)
	}

	if got := countResourceType(armTmpl.Resources, "Microsoft.ManagedIdentity/userAssignedIdentities"); got != 0 {
		t.Errorf("expected no user-assigned identities without azure roles, got %d", got)
	}

	tmplBytes, err := json.MarshalIndent(armTmpl, "", "  ")
	if err != nil {
		t.Fatalf("unable to marshal ARM template: %v", err)
	}
	body := string(tmplBytes)

	// Legacy behavior: broad grants remain on the system identity.
	if !strings.Contains(body, contributorRoleGUID) {
		t.Error("legacy Contributor grant should remain when no operation identities are configured")
	}
	if !strings.Contains(body, "custom-role-deployment") {
		t.Error("legacy register-action deployment should remain when no operation identities are configured")
	}
	if strings.Contains(body, "SystemAssigned, UserAssigned") {
		t.Error("runner VMSS should be system-assigned only when no operation identities are configured")
	}
}

func TestAzureBuiltInRoleGUID(t *testing.T) {
	if got := azureBuiltInRoleGUID("Contributor"); got != contributorRoleGUID {
		t.Errorf("Contributor = %q, want %q", got, contributorRoleGUID)
	}
	// Unknown value is treated as an already-resolved GUID.
	raw := "00000000-0000-0000-0000-000000000000"
	if got := azureBuiltInRoleGUID(raw); got != raw {
		t.Errorf("passthrough GUID = %q, want %q", got, raw)
	}
}
