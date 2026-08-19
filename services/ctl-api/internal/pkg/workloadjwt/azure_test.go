package workloadjwt

import "testing"

const (
	testTenantID = "11111111-2222-3333-4444-555555555555"
	testSubID    = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
)

func azureClaims(mirID string) map[string]any {
	return map[string]any{
		"tid":       testTenantID,
		"oid":       "99999999-8888-7777-6666-555555555555",
		"xms_mirid": mirID,
	}
}

func uamiMIRID(sub, name string) string {
	return "/subscriptions/" + sub +
		"/resourcegroups/rg-1/providers/Microsoft.ManagedIdentity/userAssignedIdentities/" + name
}

func TestAzureIssuer_RejectsNonGUIDTenant(t *testing.T) {
	// A tenant is interpolated into the issuer URL, so anything but a GUID is refused
	// before it can reshape the URL.
	for _, tenant := range []string{
		"",
		"not-a-guid",
		"../../evil.com",
		"11111111-2222-3333-4444-555555555555/../..",
		"11111111-2222-3333-4444-555555555555@evil.com",
	} {
		if _, err := AzureIssuer(tenant); err == nil {
			t.Errorf("expected %q to be rejected as an issuer tenant", tenant)
		}
	}
}

func TestAzureIssuer_BuildsV1Issuer(t *testing.T) {
	got, err := AzureIssuer(testTenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// IMDS mints v1 tokens, so the issuer must be sts.windows.net, not the v2 form.
	if want := "https://sts.windows.net/" + testTenantID + "/"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseAzureManagedIdentity_UserAssigned(t *testing.T) {
	identity, err := ParseAzureManagedIdentity(azureClaims(uamiMIRID(testSubID, "inst123-phone-home")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.SubscriptionID != testSubID {
		t.Errorf("got subscription %q", identity.SubscriptionID)
	}
	if identity.Name != "inst123-phone-home" {
		t.Errorf("got name %q", identity.Name)
	}
	if identity.ResourceGroup != "rg-1" {
		t.Errorf("got resource group %q", identity.ResourceGroup)
	}
}

func TestParseAzureManagedIdentity_RejectsSystemAssigned(t *testing.T) {
	// The runner's VMSS identity is system-assigned: xms_mirid names the compute
	// resource, and its trailing segment is a VM name rather than an identity name.
	// Accepting it would let a runner token pass as a phone-home identity.
	systemAssigned := "/subscriptions/" + testSubID +
		"/resourcegroups/rg-1/providers/Microsoft.Compute/virtualMachineScaleSets/inst123-vmss"

	if _, err := ParseAzureManagedIdentity(azureClaims(systemAssigned)); err == nil {
		t.Fatal("expected a system-assigned identity to be rejected")
	}
}

func TestParseAzureManagedIdentity_RejectsMalformed(t *testing.T) {
	cases := map[string]string{
		"empty":              "",
		"too few segments":   "/subscriptions/" + testSubID + "/resourcegroups/rg-1",
		"non guid sub":       uamiMIRID("not-a-guid", "inst123-phone-home"),
		"wrong provider":     "/subscriptions/" + testSubID + "/resourcegroups/rg-1/providers/Microsoft.Storage/userAssignedIdentities/x",
		"empty name":         uamiMIRID(testSubID, ""),
		"trailing extra seg": uamiMIRID(testSubID, "inst123-phone-home") + "/extra",
	}

	for name, mirID := range cases {
		if _, err := ParseAzureManagedIdentity(azureClaims(mirID)); err == nil {
			t.Errorf("%s: expected rejection for %q", name, mirID)
		}
	}
}

func TestParseAzureManagedIdentity_RequiresClaims(t *testing.T) {
	mirID := uamiMIRID(testSubID, "inst123-phone-home")

	for _, missing := range []string{"tid", "oid", "xms_mirid"} {
		claims := azureClaims(mirID)
		delete(claims, missing)
		if _, err := ParseAzureManagedIdentity(claims); err == nil {
			t.Errorf("expected rejection when %s is missing", missing)
		}
	}

	// A non-string claim is treated as absent rather than coerced.
	claims := azureClaims(mirID)
	claims["tid"] = 12345
	if _, err := ParseAzureManagedIdentity(claims); err == nil {
		t.Error("expected rejection for a non-string tid")
	}
}

func TestParseAzureManagedIdentity_RejectsNonGUIDTenant(t *testing.T) {
	claims := azureClaims(uamiMIRID(testSubID, "inst123-phone-home"))
	claims["tid"] = "not-a-guid"

	if _, err := ParseAzureManagedIdentity(claims); err == nil {
		t.Fatal("expected a non-guid tid to be rejected")
	}
}

func TestParseAzureManagedIdentity_CaseInsensitiveResourceSegments(t *testing.T) {
	// Azure does not preserve casing consistently in xms_mirid.
	mirID := "/SUBSCRIPTIONS/" + testSubID +
		"/RESOURCEGROUPS/rg-1/PROVIDERS/microsoft.managedidentity/USERASSIGNEDIDENTITIES/inst123-phone-home"

	identity, err := ParseAzureManagedIdentity(azureClaims(mirID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.Name != "inst123-phone-home" {
		t.Errorf("got name %q", identity.Name)
	}
}

func TestAzurePhoneHomeIdentityName(t *testing.T) {
	// Both the ARM renderer and the verifier derive the name from this, so it is the
	// single point of agreement between them.
	if got := AzurePhoneHomeIdentityName("inst123"); got != "inst123-phone-home" {
		t.Errorf("got %q", got)
	}
}

func TestAzureGraphAudience_IsNotARM(t *testing.T) {
	// Sharing the runner-auth audience would make the two credentials interchangeable.
	if AzureGraphAudience == "https://management.azure.com/" {
		t.Error("phone home must not accept the ARM audience runner auth uses")
	}
}
