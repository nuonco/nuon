package airgap

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// componentOutputPlaceholderPrefix marks tokens that plan compilation bakes
// into composite plans in place of cross-component output references. The
// runner substitutes the producing component's actual terraform outputs at
// render time; any token left after substitution is a hard error rather than
// a silently unrendered reference.
const componentOutputPlaceholderPrefix = "__NUON_AIRGAP_COMPONENT_"

// ComponentOutputPlaceholder returns the token that stands in for one
// component output reference. The trailing hash disambiguates paths whose
// sanitized forms collide (a.b vs a_b).
func ComponentOutputPlaceholder(componentName, outputPath string) string {
	sum := sha256.Sum256([]byte(componentName + "\x00" + outputPath))
	sanitized := strings.NewReplacer(".", "_", "-", "_").Replace(componentName + "_" + outputPath)
	return componentOutputPlaceholderPrefix + sanitized + "_" + hex.EncodeToString(sum[:])[:8] + "__"
}

// SubstituteComponentOutputs replaces component-output placeholders inside
// every string leaf of a decoded plan. Placeholder tokens are unique, so
// plain substring replacement cannot collide with real plan content.
func SubstituteComponentOutputs(node any, values map[string]string) {
	if len(values) == 0 {
		return
	}
	replacements := make([]string, 0, len(values)*2)
	for token, value := range values {
		replacements = append(replacements, token, value)
	}
	replacer := strings.NewReplacer(replacements...)
	rewriteStringLeaves(node, func(s string) string {
		if !strings.Contains(s, componentOutputPlaceholderPrefix) {
			return s
		}
		return replacer.Replace(s)
	})
}

// UnresolvedComponentOutputs returns the bindings whose tokens are still
// present in the plan after substitution.
func UnresolvedComponentOutputs(node any, bindings []OutputBinding) []OutputBinding {
	var missing []OutputBinding
	for _, binding := range bindings {
		if containsString(node, binding.Token) {
			missing = append(missing, binding)
		}
	}
	return missing
}

// ResolveOutputPath walks a dotted path through nested output maps and
// returns the value it lands on.
func ResolveOutputPath(outputs map[string]any, path string) (any, bool) {
	var current any = outputs
	for _, key := range strings.Split(path, ".") {
		node, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = node[key]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// OutputValueString renders a resolved output value the way it would appear
// after online template rendering: scalars as plain text, composites as
// compact JSON.
func OutputValueString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case json.Number:
		return v.String(), nil
	case nil:
		return "", fmt.Errorf("output value is null")
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("encode composite output value: %w", err)
		}
		return string(encoded), nil
	}
}
