// Package labels provides a shared parser for kubectl-style label arguments
// used by `nuon installs label`, `nuon components label`, and
// `nuon actions label`.
package labels

import (
	"fmt"
	"strings"
)

// ParseArgs parses kubectl-style label args into add and remove sets.
//
//	"key=value" -> set["key"] = "value"
//	"key-"      -> remove ["key"]
//
// Returns an error on malformed args. Empty args returns empty maps and nil error
// (caller may interpret that as "no-op / list").
func ParseArgs(args []string) (set map[string]string, remove []string, err error) {
	set = map[string]string{}
	for _, a := range args {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		// Removal form: trailing dash with no '=' anywhere.
		if strings.HasSuffix(a, "-") && !strings.Contains(a, "=") {
			key := strings.TrimSuffix(a, "-")
			if key == "" {
				return nil, nil, fmt.Errorf("invalid label %q (key cannot be empty)", a)
			}
			remove = append(remove, key)
			continue
		}
		k, v, ok := strings.Cut(a, "=")
		if !ok || k == "" {
			return nil, nil, fmt.Errorf("invalid label %q (expected key=value or key-)", a)
		}
		set[k] = v
	}
	return set, remove, nil
}
