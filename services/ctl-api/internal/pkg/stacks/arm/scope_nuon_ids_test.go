package arm

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// rootOnly strips the inline templates out of a resource, leaving only what the
// root itself evaluates. A nested deployment's inline template declares its own
// parameters and is a separate evaluation context, so references inside it are not
// the root's business.
func rootOnly(resource map[string]any) map[string]any {
	out := make(map[string]any, len(resource))
	for k, v := range resource {
		if k != "properties" {
			out[k] = v
			continue
		}
		props, ok := v.(map[string]any)
		if !ok {
			out[k] = v
			continue
		}
		stripped := make(map[string]any, len(props))
		for pk, pv := range props {
			if pk == "template" {
				continue
			}
			stripped[pk] = pv
		}
		out[k] = stripped
	}
	return out
}

// At subscription scope the Nuon-managed values are variables so the portal's
// deployment form cannot offer them as editable fields. Any expression the root
// still evaluates as parameters('nuonInstallID') would reference a parameter that no
// longer exists, and ARM rejects the whole template — so this failing is a hard
// break, not a cosmetic one.
func TestGetAzureTemplate_SubscriptionScopeRootNeverReadsNuonIDsAsParameters(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(subscriptionTemplateInput())
	if err != nil {
		t.Fatalf("render at subscription scope: %v", err)
	}

	forbidden := make([]string, 0, len(nuonIDNames)+1)
	for _, name := range append(nuonIDNames, locationVarName) {
		forbidden = append(forbidden, fmt.Sprintf("parameters('%s')", name))
	}

	check := func(label string, v any) {
		blob, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal %s: %v", label, err)
		}
		for _, bad := range forbidden {
			if strings.Contains(string(blob), bad) {
				t.Errorf("%s reads %s, which is a variable at subscription scope:\n%s", label, bad, blob)
			}
		}
	}

	check("variables", armTmpl.Variables)
	check("outputs", armTmpl.Outputs)
	for i, r := range armTmpl.Resources {
		res, ok := r.(map[string]any)
		if !ok {
			continue
		}
		check(fmt.Sprintf("resource %d (%v)", i, res["name"]), rootOnly(res))
	}
}

// The mirror of the above: inner templates must keep reading parameters, both
// because that is what they declare and because role assignment names embed the
// install ID in a guid() that must not change.
func TestGetAzureTemplate_WrappedTemplatesStillDeclareTheirOwnParameters(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(subscriptionTemplateInput())
	if err != nil {
		t.Fatalf("render at subscription scope: %v", err)
	}

	checked := 0
	for _, r := range armTmpl.Resources {
		res, ok := r.(map[string]any)
		if !ok {
			continue
		}
		props, ok := res["properties"].(map[string]any)
		if !ok {
			continue
		}
		inner, ok := props["template"].(map[string]any)
		if !ok {
			continue
		}

		declared, _ := inner["parameters"].(map[string]any)
		blob, err := json.Marshal(inner["resources"])
		if err != nil {
			t.Fatalf("marshal inner resources: %v", err)
		}

		for _, name := range append(nuonIDNames, locationVarName) {
			if !strings.Contains(string(blob), fmt.Sprintf("parameters('%s')", name)) {
				continue
			}
			if _, ok := declared[name]; !ok {
				t.Errorf("inline template of %v reads parameters('%s') without declaring it", res["name"], name)
			}
			checked++
		}
	}

	if checked == 0 {
		t.Error("no inline template read a Nuon-managed parameter; the fixture is not exercising the wrappers")
	}
}
