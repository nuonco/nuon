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

// inputDiffNode builds a diff node for a single install input. Component
// overrides render as a structured per-key YAML/HCL diff; everything else (and
// any override that fails to parse) renders as a plain string diff.
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

// structuredHelmValuesDiff builds a nested diff tree for a Helm values override
// input by parsing both the old and new YAML and diffing it key-by-key, so the
// CLI renders a per-key YAML diff instead of one opaque multi-line blob.
//
// It returns (nil, false) when either side is not a YAML mapping (e.g. a parse
// error or a top-level scalar/list), so the caller can fall back to a plain
// whole-string diff.
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

// structuredTFVarsDiff builds a nested diff tree for a Terraform vars override
// input by parsing both the old and new tfvars (HCL or JSON) and diffing it
// key-by-key, so the CLI renders a per-variable diff instead of one opaque
// multi-line blob.
//
// It returns (nil, false) when either side cannot be parsed as a flat set of
// tfvars assignments (e.g. a syntax error or an expression that references
// variables/functions), so the caller can fall back to a whole-string diff.
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

// parseYAMLMapping decodes s into a string-keyed map. An empty document decodes
// to an empty map (so a newly-set override shows every key as an addition). ok
// is false when s is not a YAML mapping.
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

// parseTFVarsMapping decodes a tfvars override (HCL or JSON) into a string-keyed
// map of its top-level variable assignments. An empty document decodes to an
// empty map. ok is false when the content cannot be parsed as a flat set of
// literal assignments.
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
		// tfvars are literal assignments, so evaluate with no variables or
		// functions in scope. Anything that needs them is not a valid tfvars
		// file and falls back to a string diff.
		val, vdiags := attr.Expr.Value(nil)
		if vdiags.HasErrors() {
			return nil, false
		}
		out[name] = ctyToInterface(val)
	}
	return out, true
}

// ctyToInterface converts a cty value (from tfvars parsing) into the same
// Go-native shape that YAML decoding produces, so both feed mappingDiff.
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
		// Preserve the exact numeric text. cty numbers are arbitrary precision,
		// so converting through int64/float64 could overflow or round and make
		// distinct values diff as equal.
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

// mappingDiff diffs two decoded mappings into a sorted list of diff nodes.
// Nested mappings recurse into child branches; everything else is compared as a
// scalar leaf.
func mappingDiff(oldMap, newMap map[string]interface{}) []*diff.Diff {
	keys := sortedUnionKeys(oldMap, newMap)
	children := make([]*diff.Diff, 0, len(keys))

	for _, k := range keys {
		oldVal, oldOK := oldMap[k]
		newVal, newOK := newMap[k]

		oldChild, oldIsMap := oldVal.(map[string]interface{})
		newChild, newIsMap := newVal.(map[string]interface{})

		// Recurse when both present sides are mappings (treating an absent side
		// as an empty mapping). A type change between scalar and mapping falls
		// through to a scalar leaf diff.
		recurse := (oldIsMap || !oldOK) && (newIsMap || !newOK) && (oldIsMap || newIsMap)
		if recurse {
			if oldChild == nil {
				oldChild = map[string]interface{}{}
			}
			if newChild == nil {
				newChild = map[string]interface{}{}
			}
			childDiffs := mappingDiff(oldChild, newChild)
			// An empty mapping appearing or disappearing has no children to
			// diff, so emit a scalar leaf to record the add/remove.
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

// scalarToString renders a leaf YAML value for display. Lists and (type-changed)
// mappings are rendered as compact JSON so they stay on a single diff line.
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
