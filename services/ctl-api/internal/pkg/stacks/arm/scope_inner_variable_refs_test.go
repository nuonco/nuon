package arm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

var (
	variableRefRegexp  = regexp.MustCompile(`variables\('([^']+)'\)`)
	parameterRefRegexp = regexp.MustCompile(`parameters\('([^']+)'\)`)
)

// A nested deployment using inner expression evaluation cannot see the root, and
// the root cannot see into it. So every variables()/parameters() read has to be
// satisfied by the declarations of the template it physically sits in, or the
// deploy fails with "The template variable 'x' is not found" — at deploy time, long
// after the render looked fine.
//
// Subscription scope makes this easy to get wrong in both directions, because it
// moves nuonInstallID, location and the Nuon IDs from root parameters to root
// variables while the wrappers still receive them as parameters.
func TestScopedExpressionsResolveWhereTheyLand(t *testing.T) {
	for _, tc := range []struct {
		name string
		inp  *stacks.TemplateInput
	}{
		{"resource group scope", azureRolesTemplateInput()},
		{"subscription scope", subscriptionTemplateInput()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := &Templates{cfg: &internal.Config{}}
			armTmpl, err := tmpl.getAzureTemplate(tc.inp)
			if err != nil {
				t.Fatalf("render: %v", err)
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
		})
	}
}

// unresolvedScopedRefs returns a sorted description of every variables()/parameters()
// read in a rendered template that the template it sits in does not declare.
func unresolvedScopedRefs(root map[string]any) []string {
	var problems []string
	for _, lvl := range append([]scopeLevel{{path: "root", body: root}}, inlineTemplates(root, "")...) {
		vars := keysOf(lvl.body["variables"])
		params := keysOf(lvl.body["parameters"])

		for ref, where := range refsIn(lvl.body, lvl.path, variableRefRegexp) {
			if vars[ref] {
				continue
			}
			problems = append(problems, describe(where, "variables", ref, params[ref], "parameter"))
		}
		for ref, where := range refsIn(lvl.body, lvl.path, parameterRefRegexp) {
			if params[ref] {
				continue
			}
			problems = append(problems, describe(where, "parameters", ref, vars[ref], "variable"))
		}
	}
	sort.Strings(problems)
	return problems
}

func describe(where, kind, ref string, declaredAsOther bool, other string) string {
	msg := fmt.Sprintf("%s reads %s('%s'), which this template does not declare", where, kind, ref)
	if declaredAsOther {
		msg += fmt.Sprintf(" (it is declared as a %s here)", other)
	}
	return msg
}

type scopeLevel struct {
	path string
	body map[string]any
}

// inlineTemplates finds every properties.template in the tree, including nested
// ones, so a wrapper inside a wrapper is checked too.
func inlineTemplates(node any, path string) []scopeLevel {
	var found []scopeLevel
	switch v := node.(type) {
	case map[string]any:
		for key, child := range v {
			p := path + "." + key
			if key == "properties" {
				if props, ok := child.(map[string]any); ok {
					if body, ok := props["template"].(map[string]any); ok {
						found = append(found, scopeLevel{path: p + ".template", body: body})
					}
				}
			}
			found = append(found, inlineTemplates(child, p)...)
		}
	case []any:
		for i, child := range v {
			found = append(found, inlineTemplates(child, fmt.Sprintf("%s[%d]", path, i))...)
		}
	}
	return found
}

// refsIn collects reads belonging to one template, skipping any deeper
// properties.template: that is its own scope and gets its own pass. A wrapper's
// name/location/dependsOn and the parameter values it passes down still belong to
// the enclosing scope, so those are kept.
func refsIn(node any, path string, re *regexp.Regexp) map[string]string {
	refs := map[string]string{}
	switch v := node.(type) {
	case string:
		for _, m := range re.FindAllStringSubmatch(v, -1) {
			if _, seen := refs[m[1]]; !seen {
				refs[m[1]] = path
			}
		}
	case map[string]any:
		for key, child := range v {
			if key == "properties" {
				if props, ok := child.(map[string]any); ok && props["template"] != nil {
					for k, c := range props {
						if k == "template" {
							continue
						}
						merge(refs, refsIn(c, path+".properties."+k, re))
					}
					continue
				}
			}
			merge(refs, refsIn(child, path+"."+key, re))
		}
	case []any:
		for i, child := range v {
			merge(refs, refsIn(child, fmt.Sprintf("%s[%d]", path, i), re))
		}
	}
	return refs
}

func merge(dst, src map[string]string) {
	for k, v := range src {
		if _, seen := dst[k]; !seen {
			dst[k] = v
		}
	}
}

func keysOf(node any) map[string]bool {
	out := map[string]bool{}
	if m, ok := node.(map[string]any); ok {
		for k := range m {
			out[k] = true
		}
	}
	return out
}
