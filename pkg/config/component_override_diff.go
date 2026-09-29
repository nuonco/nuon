package config

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	yaml "gopkg.in/yaml.v3"

	"github.com/nuonco/nuon/pkg/config/diff"
)

func inputDiffNode(key, current, val string) *diff.Diff {
	if kind, _, isOverride := ParseComponentOverrideInputName(key); isOverride {
		var (
			d  *diff.Diff
			ok bool
		)
		switch kind {
		case ComponentOverrideKindHelmValues:
			d, ok = structuredHelmValuesDiff(key, current, val)
		case ComponentOverrideKindTFVars:
			d, ok = structuredTFVarsDiff(key, current, val)
		}
		if ok {
			return d
		}
	}

	return diff.NewDiff(diff.WithKey(key), diff.WithStringDiff(current, val))
}

func structuredHelmValuesDiff(key, oldYAML, newYAML string) (*diff.Diff, bool) {
	oldMap, ok := parseYAMLMapping(oldYAML)
	if !ok {
		return nil, false
	}
	newMap, ok := parseYAMLMapping(newYAML)
	if !ok {
		return nil, false
	}

	return diff.NewDiff(
		diff.WithKey(key),
		diff.WithChildren(mappingDiff(oldMap, newMap)...),
	), true
}

func structuredTFVarsDiff(key, oldVars, newVars string) (*diff.Diff, bool) {
	oldMap, ok := parseTFVarsMapping(oldVars)
	if !ok {
		return nil, false
	}
	newMap, ok := parseTFVarsMapping(newVars)
	if !ok {
		return nil, false
	}

	return diff.NewDiff(
		diff.WithKey(key),
		diff.WithChildren(mappingDiff(oldMap, newMap)...),
	), true
}

func parseYAMLMapping(s string) (map[string]interface{}, bool) {
	if strings.TrimSpace(s) == "" {
		return map[string]interface{}{}, true
	}

	var v interface{}
	if err := yaml.Unmarshal([]byte(s), &v); err != nil {
		return nil, false
	}
	if v == nil {
		return map[string]interface{}{}, true
	}

	m, ok := v.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return m, true
}

func parseTFVarsMapping(s string) (map[string]interface{}, bool) {
	if strings.TrimSpace(s) == "" {
		return map[string]interface{}{}, true
	}

	parser := hclparse.NewParser()
	var (
		file  *hcl.File
		diags hcl.Diagnostics
	)
	if strings.HasPrefix(strings.TrimSpace(s), "{") {
		file, diags = parser.ParseJSON([]byte(s), "override.tfvars.json")
	} else {
		file, diags = parser.ParseHCL([]byte(s), "override.tfvars")
	}
	if diags.HasErrors() || file == nil {
		return nil, false
	}

	attrs, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, false
	}

	out := make(map[string]interface{}, len(attrs))
	for name, attr := range attrs {
		val, vdiags := attr.Expr.Value(nil)
		if vdiags.HasErrors() {
			return nil, false
		}
		out[name] = ctyToInterface(val)
	}
	return out, true
}

func ctyToInterface(v cty.Value) interface{} {
	if v.IsNull() || !v.IsKnown() {
		return nil
	}

	t := v.Type()
	switch {
	case t == cty.String:
		return v.AsString()
	case t == cty.Bool:
		return v.True()
	case t == cty.Number:
		return json.Number(v.AsBigFloat().Text('f', -1))
	case t.IsObjectType() || t.IsMapType():
		out := map[string]interface{}{}
		for it := v.ElementIterator(); it.Next(); {
			k, ev := it.Element()
			out[k.AsString()] = ctyToInterface(ev)
		}
		return out
	case t.IsTupleType() || t.IsListType() || t.IsSetType():
		out := []interface{}{}
		for it := v.ElementIterator(); it.Next(); {
			_, ev := it.Element()
			out = append(out, ctyToInterface(ev))
		}
		return out
	default:
		return v.GoString()
	}
}

func mappingDiff(oldMap, newMap map[string]interface{}) []*diff.Diff {
	keys := sortedUnionKeys(oldMap, newMap)
	children := make([]*diff.Diff, 0, len(keys))

	for _, k := range keys {
		oldVal, oldOK := oldMap[k]
		newVal, newOK := newMap[k]

		oldChild, oldIsMap := oldVal.(map[string]interface{})
		newChild, newIsMap := newVal.(map[string]interface{})

		recurse := (oldIsMap || !oldOK) && (newIsMap || !newOK) && (oldIsMap || newIsMap)
		if recurse {
			if oldChild == nil {
				oldChild = map[string]interface{}{}
			}
			if newChild == nil {
				newChild = map[string]interface{}{}
			}
			childDiffs := mappingDiff(oldChild, newChild)
			if len(childDiffs) == 0 && oldOK != newOK {
				children = append(children, diff.NewDiff(
					diff.WithKey(k),
					diff.WithStringDiff(scalarToString(oldVal, oldOK), scalarToString(newVal, newOK)),
				))
				continue
			}
			children = append(children, diff.NewDiff(
				diff.WithKey(k),
				diff.WithChildren(childDiffs...),
			))
			continue
		}

		children = append(children, diff.NewDiff(
			diff.WithKey(k),
			diff.WithStringDiff(scalarToString(oldVal, oldOK), scalarToString(newVal, newOK)),
		))
	}

	return children
}

func sortedUnionKeys(a, b map[string]interface{}) []string {
	set := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		set[k] = struct{}{}
	}
	for k := range b {
		set[k] = struct{}{}
	}

	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func scalarToString(v interface{}, present bool) string {
	if !present {
		return ""
	}
	if v == nil {
		return "null"
	}

	switch t := v.(type) {
	case string:
		if t == "" {
			return `""`
		}
		return t
	case map[string]interface{}, []interface{}:
		if b, err := json.Marshal(t); err == nil {
			return string(b)
		}
		return fmt.Sprint(t)
	default:
		return fmt.Sprint(t)
	}
}
