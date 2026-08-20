package arm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// serveVNetTemplate stands in for a customer-hosted VNet template.
func serveVNetTemplate(t *testing.T, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func vnetInputWithTemplate(t *testing.T, scope app.StackDeploymentScope, body string) *stacks.TemplateInput {
	t.Helper()
	inp := azureRolesTemplateInput()
	inp.DeploymentScope = scope
	inp.VPCNestedStackTemplateURL = serveVNetTemplate(t, body)
	return inp
}

const hoistFixture = `{
  "$schema": "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "nuonInstallID":   {"type": "string"},
    "location":        {"type": "string"},
    "commonTags":      {"type": "object"},
    "addressSpace":    {"type": "string", "defaultValue": "10.100.0.0/22",
                        "metadata": {"description": "VNet address space."}},
    "peeringEnabled":  {"type": "bool", "defaultValue": false},
    "requiredByOwner": {"type": "string"},
    "derivedName":     {"type": "string", "defaultValue": "[concat('rg-', uniqueString(parameters('nuonInstallID')))]"}
  },
  "resources": [],
  "outputs": {}
}`

// A VNet template is nested, so it never receives a parameter file of its own.
// Anything Nuon does not supply has to reach the root to be settable at all.
func TestVNetLinkedDeployment_HoistsNonReservedParams(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := vnetInputWithTemplate(t, app.StackDeploymentScopeSubscription, hoistFixture)

	dep, hoisted, _, err := tmpl.getVNetLinkedDeployment(inp, scopeFor(inp))
	if err != nil {
		t.Fatal(err)
	}
	params := dep["properties"].(map[string]any)["parameters"].(map[string]any)

	for _, name := range []string{"addressSpace", "peeringEnabled", "requiredByOwner"} {
		if _, ok := hoisted[name]; !ok {
			t.Errorf("%q was not hoisted, so nothing can ever set it", name)
		}
		if got := params[name]; got == nil {
			t.Errorf("%q hoisted but not threaded back down", name)
		} else if want := map[string]any{"value": "[parameters('" + name + "')]"}; got.(map[string]any)["value"] != want["value"] {
			t.Errorf("%q threaded as %v, want %v", name, got, want)
		}
	}

	// Nuon owns these; surfacing them would put internal values in the portal form.
	for _, name := range []string{"nuonInstallID", "location", "commonTags"} {
		if _, ok := hoisted[name]; ok {
			t.Errorf("%q is Nuon-managed and must not be hoisted", name)
		}
		if _, ok := params[name]; !ok {
			t.Errorf("%q is Nuon-managed but no value was passed", name)
		}
	}

	// Metadata and defaults travel with the parameter, so the portal form keeps its
	// description and pre-filled value.
	if got := hoisted["addressSpace"]; got.DefaultValue != "10.100.0.0/22" {
		t.Errorf("default not carried up: %v", got.DefaultValue)
	} else if got.Metadata == nil || got.Metadata.Description != "VNet address space." {
		t.Errorf("description not carried up: %+v", got.Metadata)
	}

	// A non-string default has to survive as its own type.
	if got := hoisted["peeringEnabled"]; got.DefaultValue != false {
		t.Errorf("bool default became %#v", got.DefaultValue)
	}

	// requiredByOwner has no default, so the portal must ask for it rather than
	// silently deploy an empty string.
	if got := hoisted["requiredByOwner"]; got.DefaultValue != nil {
		t.Errorf("invented a default: %#v", got.DefaultValue)
	}
}

// An expression default is computed in the nested template's own scope. Hoisting it
// re-binds those references to the root, where at subscription scope nuonInstallID
// is a variable and not a parameter at all.
func TestVNetLinkedDeployment_LeavesExpressionDefaultsInPlace(t *testing.T) {
	for _, scope := range []app.StackDeploymentScope{
		app.StackDeploymentScopeResourceGroup,
		app.StackDeploymentScopeSubscription,
	} {
		t.Run(string(scope), func(t *testing.T) {
			tmpl := &Templates{cfg: &internal.Config{}}
			inp := vnetInputWithTemplate(t, scope, hoistFixture)

			dep, hoisted, _, err := tmpl.getVNetLinkedDeployment(inp, scopeFor(inp))
			if err != nil {
				t.Fatal(err)
			}

			if _, ok := hoisted["derivedName"]; ok {
				t.Error("hoisted an expression default; its parameters() reads would rebind to the root")
			}
			params := dep["properties"].(map[string]any)["parameters"].(map[string]any)
			if _, ok := params["derivedName"]; ok {
				t.Error("passed a value for derivedName, overriding the default it computes for itself")
			}
		})
	}
}

// The whole point of hoisting is that the root declares what it references.
func TestVNetLinkedDeployment_HoistedParamsReachTheRoot(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := vnetInputWithTemplate(t, app.StackDeploymentScopeSubscription, hoistFixture)

	armTmpl, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"addressSpace", "peeringEnabled", "requiredByOwner"} {
		if _, ok := armTmpl.Parameters[name]; !ok {
			t.Errorf("root does not declare hoisted parameter %q", name)
		}
	}

	raw, err := json.Marshal(armTmpl)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatal(err)
	}
	for _, p := range unresolvedScopedRefs(root) {
		t.Error(p)
	}
}
