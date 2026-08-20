package arm

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

var referenceResourceIDRegexp = regexp.MustCompile(`reference\(resourceId\(`)

// ARM resolves a reference() to a resource the current template does not declare
// during preflight, before any nested deployment has run, and dependsOn does not
// defer it. So at subscription scope the root may not reference() anything that
// wrapInInstallRG relocated — the value has to come back as a nested output.
//
// The failure mode is badly misleading: the deployment reports ResourceGroupNotFound
// for a resource group that the same deployment lists as Created, because the read
// raced its creation.
func TestSubscriptionScope_RootDoesNotReferenceWrappedResources(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(subscriptionTemplateInput())
	if err != nil {
		t.Fatalf("render at subscription scope: %v", err)
	}
	raw, err := json.Marshal(armTmpl)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatal(err)
	}

	// Types the wrappers own. A root reference() to one of these is the bug.
	wrapped := map[string]bool{}
	for _, lvl := range inlineTemplates(root, "") {
		for _, r := range asSlice(lvl.body["resources"]) {
			if m, ok := r.(map[string]any); ok {
				if typ, ok := m["type"].(string); ok {
					wrapped[strings.ToLower(typ)] = true
				}
			}
		}
	}
	if len(wrapped) == 0 {
		t.Fatal("found no wrapped resources; the fixture is not exercising the wrappers")
	}

	var problems []string
	for expr, where := range rootExpressions(root, "") {
		if !referenceResourceIDRegexp.MatchString(expr) {
			continue
		}
		for typ := range wrapped {
			if strings.Contains(strings.ToLower(expr), "'"+typ+"'") {
				problems = append(problems, where+" reference()s "+typ+
					", which is declared inside a nested deployment; read it from that deployment's outputs instead")
			}
		}
	}
	sort.Strings(problems)
	for _, p := range problems {
		t.Error(p)
	}
}

// rootExpressions returns the strings belonging to the root template, skipping the
// bodies of nested inline templates.
func rootExpressions(node any, path string) map[string]string {
	out := map[string]string{}
	switch v := node.(type) {
	case string:
		out[v] = path
	case map[string]any:
		for key, child := range v {
			if key == "properties" {
				if props, ok := child.(map[string]any); ok && props["template"] != nil {
					for k, c := range props {
						if k == "template" {
							continue
						}
						for e, p := range rootExpressions(c, path+".properties."+k) {
							out[e] = p
						}
					}
					continue
				}
			}
			for e, p := range rootExpressions(child, path+"."+key) {
				out[e] = p
			}
		}
	case []any:
		for i, child := range v {
			for e, p := range rootExpressions(child, path+"[]") {
				_ = i
				out[e] = p
			}
		}
	}
	return out
}

func asSlice(v any) []any {
	s, _ := v.([]any)
	return s
}
