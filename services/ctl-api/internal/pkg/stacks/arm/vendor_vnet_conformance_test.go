package arm

import (
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// Serves the on-disk vendor VNet template so the real renderer, not a hand-copy
// of the contract, decides whether it conforms.
//
// Run with -count=1: the template is an input Go's test cache does not track, so a
// re-run after editing it otherwise replays the previous result.
func TestVendorVNetTemplateConforms(t *testing.T) {
	path := os.Getenv("VENDOR_VNET_TEMPLATE")
	if path == "" {
		t.Skip("VENDOR_VNET_TEMPLATE not set")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	inp := subscriptionTemplateInput()
	inp.VPCNestedStackTemplateURL = srv.URL

	tmpl := &Templates{cfg: &internal.Config{}}
	scope := scopeFor(inp)

	dep, hoisted, extraOutputs, err := tmpl.getVNetLinkedDeployment(inp, scope)
	if err != nil {
		t.Fatalf("render vnet deployment: %v", err)
	}
	t.Logf("passed through as install_stack outputs:")
	for _, k := range extraOutputs {
		t.Logf("    vnet_%s", snakeCase(k))
	}

	params := dep["properties"].(map[string]any)["parameters"].(map[string]any)
	for _, name := range []string{"nuonInstallID", "location", "commonTags"} {
		if _, ok := params[name]; !ok {
			t.Errorf("Nuon does not pass %q — template must declare it to receive it", name)
		}
	}
	t.Logf("passed params: %v", sortedKeys(params))
	t.Logf("hoisted to portal form: %v", sortedKeys(hoisted))

	if _, ok := dep["resourceGroup"]; ok {
		t.Error("custom vnet deployment must not be resource-group targeted")
	}
	if dep["location"] == nil {
		t.Error("subscription-targeted deployment needs a location")
	}

	// The root reads these off vnetDeployment; a missing one fails at deploy.
	var shape struct {
		Outputs map[string]json.RawMessage `json:"outputs"`
	}
	if err := json.Unmarshal(body, &shape); err != nil {
		t.Fatal(err)
	}
	full, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatalf("render root: %v", err)
	}
	rendered, err := json.Marshal(full)
	if err != nil {
		t.Fatal(err)
	}

	// Every parameter hoisted out of the VNet template has to be declared in the
	// root that now references it.
	var root map[string]any
	if err := json.Unmarshal(rendered, &root); err != nil {
		t.Fatal(err)
	}
	for _, p := range unresolvedScopedRefs(root) {
		t.Error(p)
	}
	read := map[string]bool{}
	for _, m := range regexp.MustCompile(`reference\('vnetDeployment'\)\.outputs\.([A-Za-z0-9_]+)`).FindAllStringSubmatch(string(rendered), -1) {
		read[m[1]] = true
	}
	if len(read) == 0 {
		t.Fatal("found no vnetDeployment output reads in the rendered root")
	}
	var missing []string
	for _, name := range slices.Sorted(maps.Keys(read)) {
		if _, ok := shape.Outputs[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Errorf("root reads %d vnetDeployment outputs the template does not emit: %v", len(missing), missing)
	}
	t.Logf("root reads %d vnet outputs, all present", len(read))
}

func sortedKeys[V any](m map[string]V) []string {
	return slices.Sorted(maps.Keys(m))
}
