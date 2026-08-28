package customermanaged

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// InputPlaceholderPrefix marks tokens that plan export bakes into composite
// plans in place of the vendor reference install's input values. The runner
// substitutes customer-supplied values at render time; any token left after
// substitution is a hard error rather than a silently stale vendor value.
const InputPlaceholderPrefix = "__NUON_INPUT_"

// InputPlaceholder returns the token that stands in for the named install
// input inside exported composite plans.
func InputPlaceholder(name string) string {
	return InputPlaceholderPrefix + name + "__"
}

// SubstituteInputValues replaces input placeholders inside every string leaf
// of a decoded plan. Placeholder tokens are unique, so plain substring
// replacement cannot collide with real plan content.
func SubstituteInputValues(node any, values map[string]string) {
	if len(values) == 0 {
		return
	}
	replacements := make([]string, 0, len(values)*2)
	for name, value := range values {
		replacements = append(replacements, InputPlaceholder(name), value)
	}
	replacer := strings.NewReplacer(replacements...)
	rewriteStringLeaves(node, func(s string) string {
		if !strings.Contains(s, InputPlaceholderPrefix) {
			return s
		}
		return replacer.Replace(s)
	})
}

// UnresolvedInputPlaceholders returns the names of inputs whose placeholders
// are still present in the plan, in spec order.
func UnresolvedInputPlaceholders(node any, specs []InputSpec) []string {
	var missing []string
	for _, spec := range specs {
		if containsString(node, InputPlaceholder(spec.Name)) {
			missing = append(missing, spec.Name)
		}
	}
	return missing
}

func rewriteStringLeaves(node any, rewrite func(string) string) {
	switch v := node.(type) {
	case map[string]any:
		for key, child := range v {
			if s, ok := child.(string); ok {
				v[key] = rewrite(s)
				continue
			}
			rewriteStringLeaves(child, rewrite)
		}
	case []any:
		for i, child := range v {
			if s, ok := child.(string); ok {
				v[i] = rewrite(s)
				continue
			}
			rewriteStringLeaves(child, rewrite)
		}
	}
}

func containsString(node any, needle string) bool {
	switch v := node.(type) {
	case map[string]any:
		for _, child := range v {
			if containsString(child, needle) {
				return true
			}
		}
	case []any:
		for _, child := range v {
			if containsString(child, needle) {
				return true
			}
		}
	case string:
		return strings.Contains(v, needle)
	}
	return false
}

// ValidateInputValues checks customer-provided install input values against
// the envelope's input specs before they are persisted for the runner.
// Secrets are rejected outright: there is no offline secret path yet.
func ValidateInputValues(specs []InputSpec, values map[string]string) error {
	byName := make(map[string]InputSpec, len(specs))
	for _, spec := range specs {
		byName[spec.Name] = spec
	}

	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	var problems []string
	for _, name := range names {
		spec, ok := byName[name]
		if !ok {
			problems = append(problems, fmt.Sprintf("input %q is not declared by this bundle", name))
			continue
		}
		if spec.Secret {
			problems = append(problems, fmt.Sprintf("input %q is a secret; secrets are not supported in offline customer-managed installs", name))
			continue
		}
		if !spec.Bindable {
			problems = append(problems, fmt.Sprintf("input %q is not late-bindable in this bundle; its value was fixed when the bundle was published", name))
			continue
		}
		if err := validateInputType(spec, values[name]); err != nil {
			problems = append(problems, err.Error())
		}
	}

	for _, spec := range specs {
		if !spec.Required || !spec.Bindable || spec.Secret {
			continue
		}
		if _, ok := values[spec.Name]; ok {
			continue
		}
		if spec.Default != "" {
			continue
		}
		problems = append(problems, fmt.Sprintf("required input %q has no value and no default", spec.Name))
	}

	if len(problems) > 0 {
		return fmt.Errorf("install inputs are invalid:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}

func validateInputType(spec InputSpec, value string) error {
	switch spec.Type {
	case "bool":
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("input %q must be a boolean, got %q", spec.Name, value)
		}
	case "number":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("input %q must be a number, got %q", spec.Name, value)
		}
	}
	return nil
}

// ResolveInputValues overlays customer-provided values onto spec defaults for
// every bindable, non-secret input. The result is what the runner substitutes
// into rendered plans.
//
// An unset optional input resolves to its default even when that default is
// empty: the control plane renders unset optional inputs (such as synthetic
// component overrides) as empty strings, and the offline path must match. A
// required input with no default and no value stays unresolved so its
// placeholder surfaces as a hard error instead of a silent empty value.
func ResolveInputValues(specs []InputSpec, provided map[string]string) map[string]string {
	resolved := map[string]string{}
	for _, spec := range specs {
		if !spec.Bindable || spec.Secret {
			continue
		}
		if !spec.Required || spec.Default != "" {
			resolved[spec.Name] = spec.Default
		}
		if value, ok := provided[spec.Name]; ok {
			resolved[spec.Name] = value
		}
	}
	return resolved
}
