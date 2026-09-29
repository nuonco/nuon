package labels

import (
	"fmt"
	"strings"
)

func ParseArgs(args []string) (set map[string]string, remove []string, err error) {
	set = map[string]string{}
	for _, a := range args {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
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

func ParseKeys(args []string) ([]string, error) {
	keys := make([]string, 0, len(args))
	for _, a := range args {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if strings.Contains(a, "=") {
			return nil, fmt.Errorf("invalid label key %q (provide a bare key, not key=value)", a)
		}
		keys = append(keys, strings.TrimSuffix(a, "-"))
	}
	return keys, nil
}
