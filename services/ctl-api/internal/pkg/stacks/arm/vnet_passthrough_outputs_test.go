package arm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func toJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func declared(names ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, n := range names {
		out[n] = struct{}{}
	}
	return out
}

func TestVNetPassthroughOutputs_ExcludesTheContract(t *testing.T) {
	// A template emitting only the contract has nothing to pass through, which is
	// what keeps the built-in default VNet's render unchanged.
	if got := vnetPassthroughOutputs(declared(vnetContractOutputs...)); len(got) != 0 {
		t.Errorf("contract outputs should not be passed through, got %v", got)
	}

	got := vnetPassthroughOutputs(declared(append(vnetContractOutputs,
		"resourceGroupName", "apiserverSubnetId", "installUniqueString")...))
	want := []string{"apiserverSubnetId", "installUniqueString", "resourceGroupName"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("not sorted deterministically: got %v, want %v", got, want)
		}
	}
}

func TestSnakeCase(t *testing.T) {
	for in, want := range map[string]string{
		"resourceGroupName":                "resource_group_name",
		"apiserverSubnetId":                "apiserver_subnet_id",
		"installUniqueString":              "install_unique_string",
		"platformPrivateEndpointsSubnetId": "platform_private_endpoints_subnet_id",
		"location":                         "location",
		"vnetId":                           "vnet_id",
	} {
		if got := snakeCase(in); got != want {
			t.Errorf("snakeCase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPhoneHome_VNetPassthroughIsNamespaced(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	res := tmpl.getPhoneHomeResources(minimalTemplateInput(), nil,
		[]string{"resourceGroupName", "apiserverSubnetId"}, armScope{})
	props := res[0].(map[string]any)["properties"].(map[string]any)
	script := props["scriptContent"].(string)

	for _, want := range []string{
		`"vnet_resource_group_name": "$VNET_OUT_RESOURCEGROUPNAME"`,
		`"vnet_apiserver_subnet_id": "$VNET_OUT_APISERVERSUBNETID"`,
	} {
		if !strings.Contains(script, want) {
			t.Errorf("payload missing %s", want)
		}
	}

	// The namespace is the whole point: a VNet stack that creates its own resource
	// group emits resourceGroupName, and un-namespaced that would overwrite the
	// install resource group the rest of the platform depends on.
	if strings.Contains(script, `"resource_group_name": "$VNET_OUT_RESOURCEGROUPNAME"`) {
		t.Error("vnet output overwrote the install resource_group_name")
	}
	if !strings.Contains(script, `"resource_group_name": "$RESOURCE_GROUP_NAME"`) {
		t.Error("install resource_group_name no longer reported")
	}

	byName := map[string]any{}
	for _, env := range props["environmentVariables"].([]map[string]any) {
		byName[env["name"].(string)] = env["value"]
	}
	if got := byName["VNET_OUT_RESOURCEGROUPNAME"]; got != "[string(reference('vnetDeployment').outputs.resourceGroupName.value)]" {
		t.Errorf("unexpected env value: %v", got)
	}
	if _, ok := byName["VNET_OUT_APISERVERSUBNETID"]; !ok {
		t.Error("VNET_OUT_APISERVERSUBNETID env var not declared")
	}
}

// The wrapper crosses an inner-evaluation boundary, so a reference() the root can
// resolve has to be passed in rather than re-evaluated inside.
func TestPhoneHome_VNetPassthroughSurvivesSubscriptionWrapper(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	res := tmpl.getPhoneHomeResources(subscriptionTemplateInput(), nil,
		[]string{"resourceGroupName"}, armScope{subscription: true})

	var found bool
	for _, r := range res {
		body, ok := r.(map[string]any)
		if !ok {
			continue
		}
		props, ok := body["properties"].(map[string]any)
		if !ok {
			continue
		}
		if params, ok := props["parameters"].(map[string]any); ok {
			if ev, ok := params["environmentVariables"].(map[string]any); ok {
				if strings.Contains(toJSON(t, ev), "reference('vnetDeployment').outputs.resourceGroupName") {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("passthrough reference is not evaluated at the root scope")
	}
}
