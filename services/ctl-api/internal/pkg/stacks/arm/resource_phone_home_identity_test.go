package arm

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func TestGetPhoneHomeResource_NoIdentityWhenUnset(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()

	res := tmpl.getPhoneHomeResource(inp, nil)

	if _, ok := res["identity"]; ok {
		t.Fatal("expected no identity block when phone home auth is inactive")
	}

	script := res["properties"].(map[string]any)["scriptContent"].(string)
	for _, unwanted := range []string{"169.254.169.254", "Authorization: Bearer", "-K "} {
		if strings.Contains(script, unwanted) {
			t.Errorf("script should not authenticate, but contains %q", unwanted)
		}
	}
}

func TestGetPhoneHomeResource_AttachesIdentity(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.PhoneHomeIdentityName = "inst123-phone-home"

	res := tmpl.getPhoneHomeResource(inp, nil)

	identity, ok := res["identity"].(map[string]any)
	if !ok {
		t.Fatal("expected an identity block")
	}
	if identity["type"] != "UserAssigned" {
		t.Errorf("deploymentScripts supports user-assigned identities only, got %v", identity["type"])
	}

	wantResourceID := phoneHomeIdentityResourceID("inst123-phone-home")
	assigned := identity["userAssignedIdentities"].(map[string]any)
	if _, ok := assigned[wantResourceID]; !ok {
		t.Errorf("identity %q not attached, got %v", wantResourceID, assigned)
	}

	// The script must not be able to run before the identity exists.
	var dependsOnIdentity bool
	for _, dep := range res["dependsOn"].([]string) {
		if dep == wantResourceID {
			dependsOnIdentity = true
		}
	}
	if !dependsOnIdentity {
		t.Errorf("phone home script does not depend on its identity: %v", res["dependsOn"])
	}
}

func TestGetPhoneHomeResource_AuthenticatesWithoutLeakingToken(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.PhoneHomeIdentityName = "inst123-phone-home"

	res := tmpl.getPhoneHomeResource(inp, nil)
	props := res["properties"].(map[string]any)
	script := props["scriptContent"].(string)

	for _, want := range []string{
		"169.254.169.254",
		"management.azure.com",
		`-K "$CURL_CONFIG"`,
		"chmod 600",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("script missing %q", want)
		}
	}

	// az login enumerates subscriptions and fails for a role-less identity.
	if strings.Contains(script, "az login") {
		t.Error("script should fetch the token from IMDS, not via az login")
	}

	// Deployment script logs are readable by any reader on the resource group.
	if strings.Contains(script, "echo $TOKEN") || strings.Contains(script, "set -x") {
		t.Error("script must not echo the token")
	}

	// The IMDS response is JSON, so its spacing is not part of the contract. An
	// extraction anchored on `"access_token":"` silently yields an empty token against a
	// pretty-printed response, and the script then fails the deployment.
	if !strings.Contains(script, `"access_token"[[:space:]]*:[[:space:]]*"`) {
		t.Error("token extraction must tolerate whitespace around the colon")
	}

	// A token it could not fetch must fail the deployment, never phone home unauthenticated.
	if !strings.Contains(script, "failed to acquire managed identity token") ||
		!strings.Contains(script, "exit 1") {
		t.Error("script must exit non-zero when no token could be acquired")
	}

	var found bool
	for _, env := range props["environmentVariables"].([]map[string]any) {
		if env["name"] == "PHONE_HOME_IDENTITY_CLIENT_ID" {
			found = true
			if env["value"] != phoneHomeIdentityClientID("inst123-phone-home") {
				t.Errorf("unexpected client id expression: %v", env["value"])
			}
		}
	}
	if !found {
		t.Error("PHONE_HOME_IDENTITY_CLIENT_ID not passed to the script")
	}
}

func TestGetPhoneHomeIdentityResource_HasNoRoleAssignments(t *testing.T) {
	res := getPhoneHomeIdentityResource("inst123-phone-home")

	if res["type"] != "Microsoft.ManagedIdentity/userAssignedIdentities" {
		t.Errorf("unexpected type %v", res["type"])
	}
	if res["name"] != "inst123-phone-home" {
		t.Errorf("unexpected name %v", res["name"])
	}
	// A role-less identity is what makes a stolen token inert.
	if _, ok := res["properties"]; ok {
		t.Error("phone home identity should carry no properties, and no roles")
	}
}
