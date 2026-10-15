package arm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/pkg/config"
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
                        "allowedValues": ["10.100.0.0/22", "10.200.0.0/22"],
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
	} else if len(got.AllowedValues) != 2 || got.AllowedValues[0] != "10.100.0.0/22" || got.AllowedValues[1] != "10.200.0.0/22" {
		t.Errorf("allowed values not carried up: %#v", got.AllowedValues)
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

// derivedName's default only reads nuonInstallID, so a vendor text box does not
// promote it onto the portal form.
func TestVNetLinkedDeployment_LeavesWhollyDerivedNameWhenVendorUIBindsIt(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := vnetInputWithTemplate(t, app.StackDeploymentScopeSubscription, hoistFixture)
	inp.AppCfg.StackConfig.UIDefinition = &config.UploadedUIDefinition{
		Contents: `{
		  "$schema": "https://schema.management.azure.com/schemas/0.1.2-preview/CreateUIDefinition.MultiVm.json#",
		  "handler": "Microsoft.Azure.CreateUIDef",
		  "parameters": {
		    "basics": [{"name": "derivedName", "type": "Microsoft.Common.TextBox", "label": "Name"}],
		    "steps": [],
		    "outputs": {"derivedName": "[basics('derivedName')]"}
		  }
		}`,
	}

	dep, hoisted, _, err := tmpl.getVNetLinkedDeployment(inp, scopeFor(inp))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := hoisted["derivedName"]; ok {
		t.Fatal("hoisted derivedName; its value comes from nuonInstallID")
	}
	params := dep["properties"].(map[string]any)["parameters"].(map[string]any)
	if _, ok := params["derivedName"]; ok {
		t.Fatal("passed a value for derivedName, overriding the default it computes for itself")
	}
}

const referenceDerivedFixture = `{
  "$schema": "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "nuonInstallID": {"type": "string"},
    "location": {"type": "string"},
    "commonTags": {"type": "object"},
    "resourceGroupName": {"type": "string", "defaultValue": "[format('rg-acme-net-{0}-01', parameters('nuonInstallID'))]"},
    "virtualNetworkName": {"type": "string", "defaultValue": "[format('{0}-vnet', parameters('nuonInstallID'))]"},
    "addressSpace": {"type": "string", "defaultValue": "10.100.0.0/22"}
  },
  "resources": [],
  "outputs": {}
}`

func TestWhollyDerivedParameters_ReferenceNames(t *testing.T) {
	params := map[string]ARMParameter{
		"resourceGroupName": {
			Type:         "string",
			DefaultValue: "[format('rg-acme-net-{0}-01', parameters('nuonInstallID'))]",
		},
		"virtualNetworkName": {
			Type:         "string",
			DefaultValue: "[format('{0}-vnet', parameters('nuonInstallID'))]",
		},
		"addressSpace": {
			Type:         "string",
			DefaultValue: "10.100.0.0/22",
		},
		"prefixedSpace": {
			Type:         "string",
			DefaultValue: "[format('{0}-net', parameters('addressSpace'))]",
		},
	}
	derived := whollyDerivedParameters(params, vnetReservedParamNames)

	for _, name := range []string{"resourceGroupName", "virtualNetworkName"} {
		t.Run(name, func(t *testing.T) {
			if !derived[name] {
				t.Fatalf("%s was not classified as wholly derived", name)
			}
		})
	}
	if derived["addressSpace"] {
		t.Fatal("addressSpace is a literal default and must stay customer-facing")
	}
	if derived["prefixedSpace"] {
		t.Fatal("prefixedSpace reads addressSpace, so the customer can change it")
	}
}

func TestVNetLinkedDeployment_HidesWhollyDerivedReferenceParameters(t *testing.T) {
	for _, name := range []string{"resourceGroupName", "virtualNetworkName"} {
		t.Run(name, func(t *testing.T) {
			tmpl := &Templates{cfg: &internal.Config{}}
			inp := vnetInputWithTemplate(t, app.StackDeploymentScopeSubscription, referenceDerivedFixture)
			inp.AppCfg.StackConfig.UIDefinition = &config.UploadedUIDefinition{
				Contents: fmt.Sprintf(`{
				  "$schema": "https://schema.management.azure.com/schemas/0.1.2-preview/CreateUIDefinition.MultiVm.json#",
				  "handler": "Microsoft.Azure.CreateUIDef",
				  "parameters": {
				    "basics": [{"name": %q, "type": "Microsoft.Common.TextBox", "label": "Name"}],
				    "steps": [],
				    "outputs": {%q: "[basics('%s')]"}
				  }
				}`, name, name, name),
			}

			dep, hoisted, _, err := tmpl.getVNetLinkedDeployment(inp, scopeFor(inp))
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := hoisted[name]; ok {
				t.Fatalf("hoisted %s", name)
			}
			passed := dep["properties"].(map[string]any)["parameters"].(map[string]any)
			if _, ok := passed[name]; ok {
				t.Fatalf("passed a value for %s", name)
			}
			if _, ok := hoisted["addressSpace"]; !ok {
				t.Fatal("addressSpace was not hoisted")
			}

			_, params := renderUIDef(t, inp)
			if elementNamed(params["basics"].([]any), name) || elementNamed(params["steps"].([]any), name) {
				t.Fatalf("%s is on the portal form", name)
			}
			if _, ok := params["outputs"].(map[string]any)[name]; ok {
				t.Fatalf("outputs include %s", name)
			}
			if !elementNamed(params["basics"].([]any), "addressSpace") {
				t.Fatal("addressSpace is missing from the portal form")
			}

			wrapper := renderWrapper(t, inp)
			if _, ok := wrapper["parameters"].(map[string]any)[name]; ok {
				t.Fatalf("wrapper declares %s", name)
			}
		})
	}
}

func TestVNetLinkedDeployment_HoistsExpressionThatReadsCustomerParameter(t *testing.T) {
	const body = `{
	  "$schema": "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
	  "contentVersion": "1.0.0.0",
	  "parameters": {
	    "nuonInstallID": {"type": "string"},
	    "addressSpace": {"type": "string", "defaultValue": "10.100.0.0/22"},
	    "prefixedSpace": {"type": "string", "defaultValue": "[format('{0}-net', parameters('addressSpace'))]"}
	  },
	  "resources": [],
	  "outputs": {}
	}`
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := vnetInputWithTemplate(t, app.StackDeploymentScopeSubscription, body)
	inp.AppCfg.StackConfig.UIDefinition = &config.UploadedUIDefinition{
		Contents: `{
		  "$schema": "https://schema.management.azure.com/schemas/0.1.2-preview/CreateUIDefinition.MultiVm.json#",
		  "handler": "Microsoft.Azure.CreateUIDef",
		  "parameters": {
		    "basics": [{"name": "prefixedSpace", "type": "Microsoft.Common.TextBox", "label": "Prefixed space"}],
		    "steps": [],
		    "outputs": {"prefixedSpace": "[basics('prefixedSpace')]"}
		  }
		}`,
	}

	dep, hoisted, _, err := tmpl.getVNetLinkedDeployment(inp, scopeFor(inp))
	if err != nil {
		t.Fatal(err)
	}
	param, ok := hoisted["prefixedSpace"]
	if !ok {
		t.Fatal("prefixedSpace was not hoisted")
	}
	if param.DefaultValue != nil {
		t.Fatalf("prefixedSpace default = %#v, want the expression default dropped", param.DefaultValue)
	}
	got := dep["properties"].(map[string]any)["parameters"].(map[string]any)["prefixedSpace"]
	want := "[parameters('prefixedSpace')]"
	if got.(map[string]any)["value"] != want {
		t.Fatalf("prefixedSpace passed as %#v, want %s", got, want)
	}
}

func elementNamed(nodes []any, want string) bool {
	for _, node := range nodes {
		el, ok := node.(map[string]any)
		if !ok {
			continue
		}
		if el["name"] == want {
			return true
		}
		children, _ := el["elements"].([]any)
		if elementNamed(children, want) {
			return true
		}
	}
	return false
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
	addressSpace := root["parameters"].(map[string]any)["addressSpace"].(map[string]any)
	if got := addressSpace["allowedValues"].([]any); len(got) != 2 || got[0] != "10.100.0.0/22" || got[1] != "10.200.0.0/22" {
		t.Errorf("root allowedValues = %#v", got)
	}

	wrapper := renderWrapper(t, inp)
	wrapperAddressSpace := wrapper["parameters"].(map[string]any)["addressSpace"].(map[string]any)
	if got := wrapperAddressSpace["allowedValues"].([]any); len(got) != 2 || got[0] != "10.100.0.0/22" || got[1] != "10.200.0.0/22" {
		t.Errorf("wrapper allowedValues = %#v", got)
	}
	for _, p := range unresolvedScopedRefs(root) {
		t.Error(p)
	}
}
