// Package orgfeatures holds the deployment's forced_enabled_features set. It is
// a leaf package so the config loader can populate it and GORM hooks, which
// have no access to the config object, can read it.
package orgfeatures

import (
	"strings"
	"sync/atomic"
)

var forced atomic.Pointer[map[string]bool]

// SetForced parses the comma-separated forced_enabled_features config value.
func SetForced(csv string) {
	set := make(map[string]bool)
	for _, name := range strings.Split(csv, ",") {
		if name = strings.TrimSpace(name); name != "" {
			set[name] = true
		}
	}
	forced.Store(&set)
}

// Forced returns the flags this deployment pins on for every org.
func Forced() map[string]bool {
	set := forced.Load()
	if set == nil {
		return map[string]bool{}
	}
	return *set
}

// IsForced reports whether the flag is pinned on for every org.
func IsForced(name string) bool {
	return Forced()[name]
}
